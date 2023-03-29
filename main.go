package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

func playCmd(wavName string) *exec.Cmd {
	return exec.Command("afplay", wavName)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <text>", os.Args[0])
	}

	voicepeakPath := "/Applications/voicepeak.app/Contents/MacOS/voicepeak"
	_, err := exec.LookPath(voicepeakPath)
	if err != nil {
		log.Fatalf("Command not found: %v", err)
	}

	text := os.Args[1]
	cmd1 := exec.Command(voicepeakPath, "-s", text)
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
