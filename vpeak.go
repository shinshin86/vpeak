package vpeak

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	emotionMap = map[string]string{
		"happy": "happy=100",
		"fun":   "fun=100",
		"angry": "angry=100",
		"sad":   "sad=100",
	}
)

// Options struct holds the settings for speech generation
type Options struct {
	Narrator string
	Emotion  string
	Output   string
	Silent   bool
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
		return fmt.Errorf("voicepeak command failed: %v", err)
	}

	if !opts.Silent {
		if err := PlayAudio(output); err != nil {
			return err
		}

		// if the output is not specified, delete the generated wav file
		if output == WavName {
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

	if narrator, ok := narratorMap[opts.Narrator]; ok {
		options = append([]string{"--narrator", narrator}, options...)
	} else if opts.Narrator != "" {
		log.Fatalf("Invalid narrator option: %s", opts.Narrator)
	}

	if emotion, ok := emotionMap[opts.Emotion]; ok {
		options = append([]string{"--emotion", emotion}, options...)
	} else if opts.Emotion != "" {
		log.Fatalf("Invalid emotion option: %s", opts.Emotion)
	}

	if opts.Output != "" {
		options = append([]string{"-o", opts.Output}, options...)
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
