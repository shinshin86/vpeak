package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
)

func playCmd(wavName string) *exec.Cmd {
	return exec.Command("afplay", wavName)
}

func main() {
	narratorOpt := flag.String("n", "", "Specify the narrator")
	flag.Parse()

	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s [-n] <text>", os.Args[0])
	}

	voicepeakPath := "/Applications/voicepeak.app/Contents/MacOS/voicepeak"
	_, err := exec.LookPath(voicepeakPath)
	if err != nil {
		log.Fatalf("Command not found: %v", err)
	}

	text := flag.Args()[0]

	narratorMap := map[string]string{
		"c":  "Japanese Female Child",
		"m3": "Japanese Male 3",
		"m2": "Japanese Male 2",
		"m1": "Japanese Male 1",
		"f3": "Japanese Female 3",
		"f2": "Japanese Female 2",
		"f1": "Japanese Female 1",
	}

	options := []string{"-s", text}
	if narrator, ok := narratorMap[*narratorOpt]; ok {
		options = append([]string{"--narrator", narrator}, options...)
	} else if *narratorOpt != "" {
		log.Fatalf("Invalid narrator option: %s", *narratorOpt)
	}

	cmd1 := exec.Command(voicepeakPath, options...)
	err = cmd1.Run()
	if err != nil {
		log.Fatalf("voicepeak command failed: %v", err)
	}

	wavName := "output.wav"

	cmd2 := playCmd(wavName)
	err = cmd2.Run()
	if err != nil {
		log.Fatalf("wav file play failed: %v", err)
	}

	err = os.Remove(wavName)
	if err != nil {
		log.Fatalf("Failed to delete %s: %v", wavName, err)
	}

	fmt.Println("Commands executed successfully")
}
