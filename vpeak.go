package vpeak

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

var VoicepeakPath string

func init() {
	if runtime.GOOS == "darwin" {
		VoicepeakPath = "/Applications/voicepeak.app/Contents/MacOS/voicepeak"
	} else if runtime.GOOS == "windows" {
		VoicepeakPath = "C:\\Program Files\\VOICEPEAK\\voicepeak.exe"
	} else {
		log.Fatal("Unsupported operating system")
	}
}

const (
	WavName = "output.wav"
)

var (
	narratorMap = map[string]string{
		"f1": "Japanese Female 1",
		"f2": "Japanese Female 2",
		"f3": "Japanese Female 3",
		"m1": "Japanese Male 1",
		"m2": "Japanese Male 2",
		"m3": "Japanese Male 3",
		"c":  "Japanese Female Child",
	}
)

// Options struct holds the settings for speech generation
type Options struct {
	Narrator string
	Emotion  string
	Output   string
	Silent   bool
	Speed    *int
	Pitch    *int
}

type Emotion struct {
	Happy int
	Sad   int
	Angry int
	Fun   int
}

func (e Emotion) String() string {
	var parts []string
	if e.Happy != 0 {
		parts = append(parts, fmt.Sprintf("happy=%d", e.Happy))
	}
	if e.Fun != 0 {
		parts = append(parts, fmt.Sprintf("fun=%d", e.Fun))
	}
	if e.Angry != 0 {
		parts = append(parts, fmt.Sprintf("angry=%d", e.Angry))
	}
	if e.Sad != 0 {
		parts = append(parts, fmt.Sprintf("sad=%d", e.Sad))
	}
	return strings.Join(parts, ",")
}

func (e Emotion) IsZero() bool {
	return e.Happy == 0 && e.Sad == 0 && e.Angry == 0 && e.Fun == 0
}

// GenerateSpeech generates speech audio from the given text and options
func GenerateSpeech(text string, opts Options) error {
	options := buildOptions(text, opts)
	output := opts.Output
	if output == "" {
		output = WavName
	}

	cmd1 := vpCmd(options)
	if err := cmd1.Run(); err != nil {
		return fmt.Errorf("voicepeak command failed: %w "+
			"(check that the specified narrator and emotion names are supported by VOICEPEAK)", err)
	}

	if !opts.Silent {
		if err := PlayAudio(output); err != nil {
			return err
		}

		// if the output is not specified, delete the generated wav file
		if runtime.GOOS != "windows" && output == WavName {
			if err := os.Remove(WavName); err != nil {
				return fmt.Errorf("failed to delete %s: %v", WavName, err)
			}
		}
	}

	return nil
}

// PlayAudio plays the specified audio file
func PlayAudio(wavName string) error {
	var cmd *exec.Cmd
	if runtime.GOOS == "darwin" {
		cmd = exec.Command("afplay", wavName)
	} else if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "start", "", wavName)
	} else {
		return fmt.Errorf("unsupported operating system")
	}

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("wav file play failed: %v", err)
	}
	return nil
}

// ProcessTextFiles processes text files in a directory and generates audio files
func ProcessTextFiles(dir string, opts Options) error {
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("error reading directory: %v", err)
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".txt" {
			filePath := filepath.Join(dir, file.Name())
			content, err := readTextFile(filePath)
			if err != nil {
				log.Printf("Error reading file (%s): %v", file.Name(), err)
				continue
			}

			outputName := convertWavExt(file.Name())
			outputPath := filepath.Join(dir, outputName)

			if opts.Output != "" {
				// Override directory output
				outputPath = filepath.Join(opts.Output, outputName)
			}

			localOpts := opts
			localOpts.Output = outputPath

			if err := GenerateSpeech(content, localOpts); err != nil {
				log.Printf("Error generating speech for file (%s): %v", file.Name(), err)
				continue
			}
		}
	}

	return nil
}

// ParseEmotion validates and normalizes an emotion option string.
func ParseEmotion(s string) (Emotion, error) {
	var e Emotion
	if s == "" {
		return e, nil
	}

	for _, p := range strings.Split(s, ",") {
		kv := strings.SplitN(p, "=", 2)
		name := kv[0]

		val := 100
		if len(kv) == 2 {
			v, err := strconv.Atoi(kv[1])
			if err != nil || v < 0 || v > 100 {
				return e, fmt.Errorf("invalid value: %s", kv[1])
			}
			val = v
		}

		switch name {
		case "happy":
			e.Happy = val
		case "sad":
			e.Sad = val
		case "angry":
			e.Angry = val
		case "fun":
			e.Fun = val
		default:
			return e, fmt.Errorf("invalid emotion: %s", name)
		}
	}

	return e, nil
}

// ListNarrators returns narrator names installed in VOICEPEAK.
func ListNarrators() ([]string, error) {
	return voicepeakList("--list-narrator")
}

// ListEmotions returns emotion names available for the given narrator.
func ListEmotions(narrator string) ([]string, error) {
	narrator = resolveNarratorName(narrator)
	if narrator == "" {
		return nil, fmt.Errorf("narrator is required")
	}

	return voicepeakList("--list-emotion", narrator)
}

// ValidateEmotionExpression validates and normalizes a VOICEPEAK emotion expression.
func ValidateEmotionExpression(raw string, allowed []string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}

	allowedSet := map[string]struct{}{}
	for _, name := range allowed {
		allowedSet[name] = struct{}{}
	}

	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			return "", fmt.Errorf("empty emotion segment")
		}

		name, value, hasValue := strings.Cut(part, "=")
		name = strings.TrimSpace(name)
		if name == "" {
			return "", fmt.Errorf("empty emotion name")
		}
		if _, ok := allowedSet[name]; !ok {
			return "", fmt.Errorf("invalid emotion: %s", name)
		}

		if !hasValue {
			continue
		}
		value = strings.TrimSpace(value)
		if value == "" {
			return "", fmt.Errorf("empty emotion value for %s", name)
		}
		weight, err := strconv.Atoi(value)
		if err != nil {
			return "", fmt.Errorf("invalid emotion weight for %s: %w", name, err)
		}
		if weight < 0 || weight > 100 {
			return "", fmt.Errorf("emotion weight for %s must be between 0 and 100", name)
		}
	}

	return raw, nil
}

// normalizeEmotionExpression validates an emotion expression syntactically and
// returns it in a canonical form. Any emotion name is accepted so that character
// products with their own emotions work; whether the name is actually supported
// is left to VOICEPEAK. Bare names are expanded to "name=100" and zero-weighted
// emotions are dropped, matching the previous Emotion.String behavior.
func normalizeEmotionExpression(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", nil
	}

	var parts []string
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			return "", fmt.Errorf("empty emotion segment")
		}

		name, value, hasValue := strings.Cut(part, "=")
		name = strings.TrimSpace(name)
		if name == "" {
			return "", fmt.Errorf("empty emotion name")
		}

		weight := 100
		if hasValue {
			value = strings.TrimSpace(value)
			if value == "" {
				return "", fmt.Errorf("empty emotion value for %s", name)
			}
			w, err := strconv.Atoi(value)
			if err != nil {
				return "", fmt.Errorf("invalid emotion weight for %s: %w", name, err)
			}
			if w < 0 || w > 100 {
				return "", fmt.Errorf("emotion weight for %s must be between 0 and 100", name)
			}
			weight = w
		}

		if weight == 0 {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%d", name, weight))
	}

	return strings.Join(parts, ","), nil
}

func voicepeakList(args ...string) ([]string, error) {
	cmd := vpCmd(args)
	output, err := cmd.CombinedOutput()
	if err != nil {
		output = bytes.TrimSpace(output)
		if len(output) == 0 {
			return nil, fmt.Errorf("voicepeak command failed: %w", err)
		}
		return nil, fmt.Errorf("voicepeak command failed: %w: %s", err, output)
	}

	return parseVoicepeakListOutput(string(output)), nil
}

func parseVoicepeakListOutput(output string) []string {
	var items []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "[debug]") || strings.HasPrefix(line, "iconv_open ") {
			continue
		}
		items = append(items, line)
	}
	return items
}

func resolveNarratorName(narrator string) string {
	if resolved, ok := narratorMap[narrator]; ok {
		return resolved
	}
	return narrator
}

func vpCmd(options []string) *exec.Cmd {
	_, err := exec.LookPath(VoicepeakPath)
	if err != nil {
		log.Fatalf("Command not found: %v", err)
	}

	return exec.Command(VoicepeakPath, options...)
}

func convertWavExt(filename string) string {
	oldExt := filepath.Ext(filename)
	return strings.TrimSuffix(filename, oldExt) + ".wav"
}

func buildOptions(text string, opts Options) []string {
	options := []string{"-s", text}

	narrator := resolveNarratorName(opts.Narrator)
	if narrator != "" {
		options = append([]string{"--narrator", narrator}, options...)
	}

	if opts.Emotion != "" {
		emotion, err := normalizeEmotionExpression(opts.Emotion)
		if err != nil {
			log.Fatalf("Invalid emotion option: %s", opts.Emotion)
		}
		if emotion != "" {
			options = append([]string{"--emotion", emotion}, options...)
		}
	}

	if opts.Output != "" {
		options = append([]string{"-o", opts.Output}, options...)
	}

	if opts.Speed != nil {
		options = append(options, "--speed", fmt.Sprintf("%d", *opts.Speed))
	}

	if opts.Pitch != nil {
		options = append(options, "--pitch", fmt.Sprintf("%d", *opts.Pitch))
	}

	return options
}

func readTextFile(filePath string) (string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}
