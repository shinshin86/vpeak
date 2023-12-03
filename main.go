package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	VoicepeakPath = "/Applications/voicepeak.app/Contents/MacOS/voicepeak"
	WavName       = "output.wav"
)

var (
	dirOpt      = flag.String("d", "", "Directory to read files from")
	outputOpt   = flag.String("o", WavName, "Output file path (Specify the name of the output directory if reading by directory (-d option))")
	narratorOpt = flag.String("n", "", "Specify the narrator. See below for options.")
	emotionOpt  = flag.String("e", "", "Specify the emotion. See below for options.")
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

func playCmd(wavName string) *exec.Cmd {
	return exec.Command("afplay", wavName)
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

func processOptions() []string {
	text := flag.Args()[0]
	options := []string{"-s", text}

	narrator, ok := narratorMap[*narratorOpt]

	if !ok && *narratorOpt != "" {
		log.Fatalf("Invalid narrator option: %s", *narratorOpt)
	}
	if ok {
		options = append([]string{"--narrator", narrator}, options...)
	}

	emotion, ok := emotionMap[*emotionOpt]
	if !ok && *emotionOpt != "" {
		log.Fatalf("Invalid emotion option: %s", *emotionOpt)
	}
	if ok {
		options = append([]string{"--emotion", emotion}, options...)
	}

	if *outputOpt != "" {
		options = append([]string{"-o", *outputOpt}, options...)
	}

	return options
}

func executeCommands(options []string, output string) {
	cmd1 := vpCmd(options)
	handleError(cmd1.Run(), "voicepeak command failed")

	cmd2 := playCmd(output)
	handleError(cmd2.Run(), "wav file play failed")

	if output == WavName {
		handleError(os.Remove(WavName), fmt.Sprintf("Failed to delete %s", WavName))
	}
}

func handleError(err error, message string) {
	if err != nil {
		log.Fatalf("%s: %v", message, err)
	}
}

func readTextFile(filePath string) (string, error) {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func processTextFiles(dir string) {
	files, err := ioutil.ReadDir(dir)
	handleError(err, "Error reading directory")

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
			options := buildOptions(content, outputPath)

			executeCommands(options, outputPath)
		}
	}
}

func buildOptions(content, outputPath string) []string {
	if *outputOpt != "" {
		// Override directory output
		outputPath = filepath.Join(*outputOpt, convertWavExt(filepath.Base(outputPath)))
	}
	options := []string{"-s", content, "-o", outputPath}

	addOption := func(key string, value string) {
		if value != "" {
			options = append([]string{key, value}, options...)
		}
	}

	narrator, ok := narratorMap[*narratorOpt]
	if !ok && *narratorOpt != "" {
		log.Fatalf("Invalid narrator option: %s", *narratorOpt)
	}
	if ok {
		addOption("--narrator", narrator)
	}

	emotion, ok := emotionMap[*emotionOpt]
	if !ok && *emotionOpt != "" {
		log.Fatalf("Invalid emotion option: %s", *emotionOpt)
	}
	if ok {
		addOption("--emotion", emotion)
	}

	return options
}

func main() {
	flag.Usage = func() {
		fmt.Printf("Usage: %s [OPTIONS] <text>\n", os.Args[0])
		fmt.Println("Options:")
		flag.PrintDefaults()
		fmt.Println("\nNarrator options:")
		fmt.Println("  f1: Japanese Female 1")
		fmt.Println("  f2: Japanese Female 2")
		fmt.Println("  f3: Japanese Female 3")
		fmt.Println("  m1: Japanese Male 1")
		fmt.Println("  m2: Japanese Male 2")
		fmt.Println("  m3: Japanese Male 3")
		fmt.Println("  c:  Japanese Female Child")
		fmt.Println("\nEmotion options:")
		fmt.Println("  happy")
		fmt.Println("  fun")
		fmt.Println("  angry")
		fmt.Println("  sad")
	}

	help := flag.Bool("help", false, "Show help")

	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s [-n] <text>", os.Args[0])
	}

	if *dirOpt == "" {
		options := processOptions()
		output := WavName
		if *outputOpt != "" {
			output = *outputOpt
		}
		executeCommands(options, output)
	} else {
		processTextFiles(*dirOpt)
	}

	fmt.Println("Commands executed successfully")
}
