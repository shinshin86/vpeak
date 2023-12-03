# vpeak
CLI tool to touch [VOICEPEAK](https://www.ah-soft.com/voice/6nare/) from the command line.

## Usage

Execute the following command to speak the string passed as an argument.

```
vpeak こんにちは！

# option (narrator: Japanese Female 1, emotion: happy)
vpeak -n f1 -e happy "こんにちは"

# option (narrator: Japanese Female 1, emotion: happy, output path: ./hello.wav)
vpeak -n f1 -e happy -o ./hello.wav "こんにちは"
```

Converts all text files(`.txt`) in the directory specified by the `-d` option to audio files (`.wav`).

```
vpeak -d your-dir

# option (narrator: Japanese Female 1, emotion: happy)
vpeak -n f1 -e happy -d your-dir

# option (narrator: Japanese Female 1, emotion: happy, output dir: your-dir-2)
vpeak -n f1 -e happy -o your-dir-2 -d your-dir
```

Multiple options can be combined.

```sh
# ex: Convert a text file in the testdir directory into a voice file with the voice of Japanese Female 1.
vpeak -d testdir -n f1
```

Run the `help` command for more information.

```
vpeak -h
```

## Support
Tested only under the following conditions.

### OS
Currently only **M1 or later(arm64) mac** are supported

### VOICEPEAK
VOICEPEAK must be updated to the latest version in order to use vpeak.  
I am testing with 1.2.7.


## License
[MIT](./LICENSE)

## Author

[Yuki Shindo](https://shinshin86.com/en)