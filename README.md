# vpeak
`vpeak` is a tool that allows you to interact with [VOICEPEAK](https://www.ah-soft.com/voice/6nare/) from the command line or within your Go applications.

## Features

- **CLI Tool**: Use `vpeak` from the command line to generate speech audio.
- **Go Library**: Import `vpeak` into your Go projects to generate speech programmatically.

---
## Handling Audio Files: Platform Differences

### About audio files (.wav)

The behavior of audio file handling depends on the options provided:

- **On macOS**: Temporary `.wav` files are automatically deleted after playback unless explicitly preserved using the `-silent`, `-o`, or `-d` options.
- **On Windows**: `.wav` files are never automatically deleted after playback, ensuring compatibility with Windows' file handling.
You may need to manually delete the files after playback if you don't want them to persist.
---
## CLI Usage

### Installation

To install `vpeak` as a CLI tool, run the following command:

```sh
go install github.com/shinshin86/vpeak/cmd/vpeak@latest
```

This will download, build, and install the vpeak binary to your $GOBIN directory (usually $HOME/go/bin).

### Usage

Execute the following command to have VOICEPEAK speak the string passed as an argument:

```sh
# # not option specfied (narrator: Japanese Female Child, emotion: natural)
vpeak こんにちは！

# option (narrator: Japanese Female 1, emotion: happy)
vpeak -n f1 -e happy "こんにちは"

# option (narrator: Japanese Female 1, emotion: happy, output path: ./hello.wav)
# (An audio file will only be generated if the output option is specified, and it will be saved at the designated location.)
vpeak -n f1 -e happy -o ./hello.wav "こんにちは"
```

Converts all text files(`.txt`) in the directory specified by the `-d` option to audio files (`.wav`).

```sh
vpeak -d your-dir

# option (narrator: Japanese Female 1, emotion: happy)
vpeak -n f1 -e happy -d your-dir

# option (narrator: Japanese Female 1, emotion: happy, output dir: your-dir-2)
vpeak -n f1 -e happy -o your-dir-2 -d your-dir

# option (speed: 120, pitch: 20)
vpeak -speed 120 -pitch 20 "こんにちは"
```

### Silent mode

When the `-silent` option is used, no voice playback is performed, and the generated files are not automatically deleted. This option is useful if you only want to generate audio files.

```
vpeak -silent "こんにちは"
```

### About audio files(.wav)

The audio file will remain only if outputPath is specified, executed per directory, or silent mode is enabled.

### Command infomation

Run the `help` command for more information.

```
vpeak -h
```

---
## Library Usage

You can also use `vpeak` as a Go library in your own applications.

### Installation

To install the library, run:

```sh
go get github.com/shinshin86/vpeak@latest
```

### Importing the Library

In your Go code, import the `vpeak` package:

```go
import "github.com/shinshin86/vpeak"
```

### Example Usage

Here's an example of how to use `vpeak` in your Go program:

```go
package main

import (
    "fmt"
    "log"

    "github.com/shinshin86/vpeak"
)

func main() {
    text := "こんにちは"
    opts := vpeak.Options{
        Narrator: "f1",       // Narrator option (e.g., "f1", "m1")
        Emotion:  "happy",    // Emotion option (e.g., "happy", "sad")
        Output:   "hello.wav",// Output file path
        Silent:   false,      // Silent mode (true or false)
        // Speed and Pitch accept *int. Use local variables to set values.
        // Example:
        // speed := 120
        // pitch := 30
        // opts.Speed = &speed
        // opts.Pitch = &pitch
    }

    if err := vpeak.GenerateSpeech(text, opts); err != nil {
        log.Fatalf("Failed to generate speech: %v", err)
    }

    fmt.Println("Speech generated successfully.")
}
```

### Options

- `Narrator`: Choose the narrator's voice. Available options:
  - `f1`: Japanese Female 1
  - `f2`: Japanese Female 2
  - `f3`: Japanese Female 3
  - `m1`: Japanese Male 1
  - `m2`: Japanese Male 2
  - `m3`: Japanese Male 3
  - `c`:  Japanese Female Child
- `Emotion`: Choose the emotion. Available options:
  - `happy`
  - `fun`
  - `angry`
  - `sad`
  - If no option is specified, it will be `natural`.
- `Output`: Specify the output file path. If not set, defaults to `output.wav`.
- `Silent`: Set to `true` to disable voice playback.
- `Speed`: Adjust speech speed (50–200). Provide as `*int`; `nil` keeps the VOICEPEAK default.
- `Pitch`: Adjust pitch (-300–300). Provide as `*int`; `nil` keeps the VOICEPEAK default.

### Processing Text Files in a Directory

You can also process all text files in a directory:

```go
package main

import (
    "fmt"
    "log"

    "github.com/shinshin86/vpeak"
)

func main() {
    dir := "your-dir"
    opts := vpeak.Options{
        Narrator: "f1",
        Emotion:  "happy",
        Output:   "your-dir-2", // Output directory
        Silent:   true,
    }

    if err := vpeak.ProcessTextFiles(dir, opts); err != nil {
        log.Fatalf("Failed to process text files: %v", err)
    }

    fmt.Println("Text files processed successfully.")
}
```

---

## Support
vpeak is currently tested under the following conditions. Compatibility with other environments is not guaranteed.

### OS
- macOS: M1 or later (arm64)
- Windows: Windows 11 (64-bit)

(We have not tested outside of these versions.)

### VOICEPEAK
- Updated to the latest version (tested with `1.2.7`)
- Default paths for macOS & Windows are defined in [the code](https://github.com/shinshin86/vpeak/blob/main/vpeak.go#L16-L24).
If VOICEPEAK is installed elsewhere, update the VoicepeakPath variable.

### Other OS
- Currently, Linux and other operating systems are not supported.

## License
[MIT](./LICENSE)

## Author

[Yuki Shindo](https://shinshin86.com/en)
