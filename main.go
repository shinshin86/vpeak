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

func main() {
	dirOpt := flag.String("d", "", "Directory to read files from")
	narratorOpt := flag.String("n", "", "Specify the narrator. See below for options.")

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

	narratorMap := map[string]string{
		"f1": "Japanese Female 1",
		"f2": "Japanese Female 2",
		"f3": "Japanese Female 3",
		"m1": "Japanese Male 1",
		"m2": "Japanese Male 2",
		"m3": "Japanese Male 3",
		"c":  "Japanese Female Child",
	}

	if *dirOpt == "" {
		text := flag.Args()[0]
		options := []string{"-s", text}
		if narrator, ok := narratorMap[*narratorOpt]; ok {
			options = append([]string{"--narrator", narrator}, options...)
		} else if *narratorOpt != "" {
			log.Fatalf("Invalid narrator option: %s", *narratorOpt)
		}

		cmd1 := vpCmd(options)
		err := cmd1.Run()
		if err != nil {
			log.Fatalf("voicepeak command failed: %v", err)
		}

		cmd2 := playCmd(WavName)
		err = cmd2.Run()
		if err != nil {
			log.Fatalf("wav file play failed: %v", err)
		}

		err = os.Remove(WavName)
		if err != nil {
			log.Fatalf("Failed to delete %s: %v", WavName, err)
		}
	} else {
		files, err := ioutil.ReadDir(*dirOpt)
		if err != nil {
			log.Fatalf("Error reading directory: %v", err)
		}

		for _, file := range files {
			if !file.IsDir() && filepath.Ext(file.Name()) == ".txt" {
				filePath := filepath.Join(*dirOpt, file.Name())
				content, err := ioutil.ReadFile(filePath)

				if err != nil {
					log.Printf("Error reading file (%s): %v", file.Name(), err)
					continue
				}

				outputName := convertWavExt(file.Name())
				outputPath := filepath.Join(*dirOpt, outputName)
				options := []string{"-s", string(content), "-o", outputPath}
				if narrator, ok := narratorMap[*narratorOpt]; ok {
					options = append([]string{"--narrator", narrator}, options...)
				} else if *narratorOpt != "" {
					log.Fatalf("Invalid narrator option: %s", *narratorOpt)
				}

				cmd1 := vpCmd(options)
				err = cmd1.Run()
				if err != nil {
					log.Fatalf("voicepeak command failed: %v", err)
				}
			}
		}
	}

	fmt.Println("Commands executed successfully")
}
