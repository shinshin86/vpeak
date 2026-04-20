package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/shinshin86/vpeak"
)

var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "dict" {
		runDictCommand(os.Args[2:])
		return
	}

	runSpeakCommand(os.Args[1:])
}

func runSpeakCommand(args []string) {
	flagSet := flag.NewFlagSet("vpeak", flag.ExitOnError)

	var (
		dirOpt      = flagSet.String("d", "", "Directory to read files from")
		outputOpt   = flagSet.String("o", "", "Output file path (Specify the name of the output directory if reading by directory (-d option))")
		narratorOpt = flagSet.String("n", "", "Specify the narrator. See below for options.")
		emotionOpt  = flagSet.String("e", "", "Specify the emotion. See below for options.")
		speedOpt    = flagSet.String("speed", "", "Specify the speech speed (50-200)")
		pitchOpt    = flagSet.String("pitch", "", "Specify the pitch adjustment (-300 - 300)")
		silentOpt   = flagSet.Bool("silent", false, "Silent mode (no sound)")
		versionOpt  = flagSet.Bool("version", false, "Show version")
		helpOpt     = flagSet.Bool("help", false, "Show help")
	)

	flagSet.Usage = func() {
		fmt.Printf("Usage: %s [OPTIONS] <text>\n", os.Args[0])
		fmt.Println("Options:")
		flagSet.PrintDefaults()
		fmt.Println("\nNarrator options:")
		fmt.Println("  f1: Japanese Female 1")
		fmt.Println("  f2: Japanese Female 2")
		fmt.Println("  f3: Japanese Female 3")
		fmt.Println("  m1: Japanese Male 1")
		fmt.Println("  m2: Japanese Male 2")
		fmt.Println("  m3: Japanese Male 3")
		fmt.Println("  c:  Japanese Female Child")
		fmt.Println("\nEmotion options (values 0-100, comma-separate multiple):")
		fmt.Println("  happy")
		fmt.Println("  fun")
		fmt.Println("  angry")
		fmt.Println("  sad")
		fmt.Println("\nEmotion examples:")
		fmt.Println("  happy              (equivalent to happy=100)")
		fmt.Println("  happy=50")
		fmt.Println("  happy=40,fun=60")
		fmt.Println("\nDictionary commands:")
		fmt.Printf("  %s dict -h\n", os.Args[0])
	}

	if err := flagSet.Parse(args); err != nil {
		log.Fatalf("Error: %v", err)
	}

	if *helpOpt {
		flagSet.Usage()
		os.Exit(0)
	}

	if *versionOpt {
		fmt.Println(version)
		os.Exit(0)
	}

	if len(flagSet.Args()) == 0 && *dirOpt == "" {
		log.Fatalf("Usage: %s [-n] <text>", os.Args[0])
	}

	opts := vpeak.Options{
		Narrator: *narratorOpt,
		Emotion:  *emotionOpt,
		Output:   *outputOpt,
		Silent:   *silentOpt,
	}

	if *speedOpt != "" {
		speedVal, err := strconv.Atoi(*speedOpt)
		if err != nil {
			log.Fatalf("Invalid speed value: %v", err)
		}
		if speedVal < 50 || speedVal > 200 {
			log.Fatalf("Speed must be between 50 and 200")
		}
		speed := speedVal
		opts.Speed = &speed
	}

	if *pitchOpt != "" {
		pitchVal, err := strconv.Atoi(*pitchOpt)
		if err != nil {
			log.Fatalf("Invalid pitch value: %v", err)
		}
		if pitchVal < -300 || pitchVal > 300 {
			log.Fatalf("Pitch must be between -300 and 300")
		}
		pitch := pitchVal
		opts.Pitch = &pitch
	}

	if *dirOpt == "" {
		text := flagSet.Args()[0]
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

func runDictCommand(args []string) {
	if len(args) == 0 {
		printDictUsage()
		os.Exit(1)
	}

	switch args[0] {
	case "-h", "--help", "help":
		printDictUsage()
	case "list":
		runDictList(args[1:])
	case "add":
		runDictAdd(args[1:])
	case "update-by-surface":
		runDictUpdateBySurface(args[1:])
	case "delete-by-surface":
		runDictDeleteBySurface(args[1:])
	case "import":
		runDictImport(args[1:])
	case "export":
		runDictExport(args[1:])
	case "path":
		runDictPath(args[1:])
	default:
		log.Fatalf("Unknown dict command: %s", args[0])
	}
}

func printDictUsage() {
	fmt.Printf("Usage: %s dict <command> [options]\n", os.Args[0])
	fmt.Println("Commands:")
	fmt.Println("  list               Print the current VOICEPEAK dictionary as JSON")
	fmt.Println("  add                Add a dictionary word")
	fmt.Println("  update-by-surface  Update a dictionary word by current surface")
	fmt.Println("  delete-by-surface  Delete a dictionary word by surface")
	fmt.Println("  import             Import dictionary entries from a JSON file")
	fmt.Println("  export             Export dictionary entries to a JSON file")
	fmt.Println("  path               Print the default dictionary path")
}

func runDictList(args []string) {
	flagSet := flag.NewFlagSet("dict list", flag.ExitOnError)
	fileOpt := flagSet.String("file", "", "Dictionary file path")
	if err := flagSet.Parse(args); err != nil {
		log.Fatalf("Error: %v", err)
	}

	path := resolveDictionaryPath(*fileOpt)
	entries, err := vpeak.LoadDictionary(path)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	fmt.Println(string(data))
}

func runDictAdd(args []string) {
	flagSet := flag.NewFlagSet("dict add", flag.ExitOnError)
	fileOpt := flagSet.String("file", "", "Dictionary file path")
	surfaceOpt := flagSet.String("surface", "", "Surface form")
	pronunciationOpt := flagSet.String("pronunciation", "", "Pronunciation in katakana")
	posOpt := flagSet.String("pos", "Japanese_Koyuumeishi_ippan", "Dictionary part-of-speech")
	priorityOpt := flagSet.Int("priority", 5, "Priority (0-10)")
	accentTypeOpt := flagSet.Int("accent-type", 0, "Accent type")
	langOpt := flagSet.String("lang", "ja", "Language code")
	if err := flagSet.Parse(args); err != nil {
		log.Fatalf("Error: %v", err)
	}

	entry := vpeak.DictEntry{
		Surface:       *surfaceOpt,
		Pronunciation: *pronunciationOpt,
		Pos:           *posOpt,
		Priority:      *priorityOpt,
		AccentType:    *accentTypeOpt,
		Lang:          *langOpt,
	}

	if err := vpeak.AddDictionaryWord(resolveDictionaryPath(*fileOpt), entry); err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Dictionary word added successfully")
}

func runDictUpdateBySurface(args []string) {
	flagSet := flag.NewFlagSet("dict update-by-surface", flag.ExitOnError)
	fileOpt := flagSet.String("file", "", "Dictionary file path")
	currentSurfaceOpt := flagSet.String("current-surface", "", "Current surface form")
	surfaceOpt := flagSet.String("surface", "", "New surface form")
	pronunciationOpt := flagSet.String("pronunciation", "", "Pronunciation in katakana")
	posOpt := flagSet.String("pos", "Japanese_Koyuumeishi_ippan", "Dictionary part-of-speech")
	priorityOpt := flagSet.Int("priority", 5, "Priority (0-10)")
	accentTypeOpt := flagSet.Int("accent-type", 0, "Accent type")
	langOpt := flagSet.String("lang", "ja", "Language code")
	if err := flagSet.Parse(args); err != nil {
		log.Fatalf("Error: %v", err)
	}

	entry := vpeak.DictEntry{
		Surface:       *surfaceOpt,
		Pronunciation: *pronunciationOpt,
		Pos:           *posOpt,
		Priority:      *priorityOpt,
		AccentType:    *accentTypeOpt,
		Lang:          *langOpt,
	}

	if err := vpeak.UpdateDictionaryWordBySurface(resolveDictionaryPath(*fileOpt), *currentSurfaceOpt, entry); err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Dictionary word updated successfully")
}

func runDictDeleteBySurface(args []string) {
	flagSet := flag.NewFlagSet("dict delete-by-surface", flag.ExitOnError)
	fileOpt := flagSet.String("file", "", "Dictionary file path")
	surfaceOpt := flagSet.String("surface", "", "Surface form")
	if err := flagSet.Parse(args); err != nil {
		log.Fatalf("Error: %v", err)
	}

	if err := vpeak.DeleteDictionaryWordBySurface(resolveDictionaryPath(*fileOpt), *surfaceOpt); err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Dictionary word deleted successfully")
}

func runDictImport(args []string) {
	flagSet := flag.NewFlagSet("dict import", flag.ExitOnError)
	fileOpt := flagSet.String("file", "", "Dictionary file path")
	importFileOpt := flagSet.String("import-file", "", "Dictionary JSON file to import")
	overrideOpt := flagSet.Bool("override", false, "Override existing entries matched by surface")
	if err := flagSet.Parse(args); err != nil {
		log.Fatalf("Error: %v", err)
	}

	if *importFileOpt == "" {
		log.Fatalf("Error: import-file is required")
	}

	importedEntries, err := vpeak.LoadDictionary(*importFileOpt)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	if err := vpeak.ImportDictionary(resolveDictionaryPath(*fileOpt), importedEntries, *overrideOpt); err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Dictionary imported successfully")
}

func runDictExport(args []string) {
	flagSet := flag.NewFlagSet("dict export", flag.ExitOnError)
	fileOpt := flagSet.String("file", "", "Dictionary file path")
	exportFileOpt := flagSet.String("export-file", "", "Export destination path")
	if err := flagSet.Parse(args); err != nil {
		log.Fatalf("Error: %v", err)
	}

	if *exportFileOpt == "" {
		log.Fatalf("Error: export-file is required")
	}

	if err := vpeak.ExportDictionary(resolveDictionaryPath(*fileOpt), *exportFileOpt); err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println("Dictionary exported successfully")
}

func runDictPath(args []string) {
	flagSet := flag.NewFlagSet("dict path", flag.ExitOnError)
	if err := flagSet.Parse(args); err != nil {
		log.Fatalf("Error: %v", err)
	}

	path, err := vpeak.DefaultDictionaryPath()
	if err != nil {
		log.Fatalf("Error: %v", err)
	}

	fmt.Println(path)
}

func resolveDictionaryPath(path string) string {
	if path != "" {
		return path
	}

	defaultPath, err := vpeak.DefaultDictionaryPath()
	if err != nil {
		if errors.Is(err, vpeak.ErrDictionaryPathUnsupported) {
			log.Fatalf("Error: %v", err)
		}
		log.Fatalf("Error: %v", err)
	}

	return defaultPath
}
