package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/shinshin86/vpeak"
)

func main() {
	var (
		dirOpt      = flag.String("d", "", "Directory to read files from")
		outputOpt   = flag.String("o", "", "Output file path (Specify the name of the output directory if reading by directory (-d option))")
		narratorOpt = flag.String("n", "", "Specify the narrator. See below for options.")
		emotionOpt  = flag.String("e", "", "Specify the emotion. See below for options.")
		silentOpt   = flag.Bool("silent", false, "Silent mode (no sound)")
	)

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

	if len(flag.Args()) == 0 && *dirOpt == "" {
		log.Fatalf("Usage: %s [-n] <text>", os.Args[0])
	}

	opts := vpeak.Options{
		Narrator: *narratorOpt,
		Emotion:  *emotionOpt,
		Output:   *outputOpt,
		Silent:   *silentOpt,
	}

	if *dirOpt == "" {
		text := flag.Args()[0]
		if err := vpeak.GenerateSpeech(text, opts); err != nil {
			log.Fatalf("Error: %v", err)
		}
	} else {
		if err := vpeak.ProcessTextFiles(*dirOpt, opts); err != nil {
			log.Fatalf("Error: %v", err)
		}
	}

	fmt.Println("Commands executed successfully")
}
