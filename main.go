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

	voicepeakPath := "/Applications/voicepeak.app/Contents/MacOS/voicepeak"
	_, err := exec.LookPath(voicepeakPath)
	if err != nil {
		log.Fatalf("Command not found: %v", err)
	}

	text := flag.Args()[0]

	narratorMap := map[string]string{
		"f1": "Japanese Female 1",
		"f2": "Japanese Female 2",
		"f3": "Japanese Female 3",
		"m1": "Japanese Male 1",
		"m2": "Japanese Male 2",
		"m3": "Japanese Male 3",
		"c":  "Japanese Female Child",
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
