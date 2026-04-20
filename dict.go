package vpeak

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

const defaultDictionaryLang = "ja"

var (
	ErrDictionaryWordNotFound    = errors.New("dictionary word not found")
	ErrDictionaryWordConflict    = errors.New("dictionary word conflict")
	ErrDictionaryWordInvalid     = errors.New("dictionary word invalid")
	ErrDictionaryPathUnsupported = errors.New("dictionary path unsupported")
)

var (
	validDictionaryPos = map[string]bool{
		"Japanese_Futsuu_meishi":      true,
		"Japanese_Koyuumeishi_ippan":  true,
		"Japanese_Koyuumeishi_jinmei": true,
		"Japanese_Koyuumeishi_mei":    true,
		"Japanese_Koyuumeishi_place":  true,
		"Japanese_Koyuumeishi_sei":    true,
	}
	katakanaPattern = regexp.MustCompile(`^[ァ-ヶー]+$`)
)

type DictEntry struct {
	Surface       string `json:"sur"`
	Pronunciation string `json:"pron"`
	Pos           string `json:"pos"`
	Priority      int    `json:"priority"`
	AccentType    int    `json:"accentType"`
	Lang          string `json:"lang"`
}

func DefaultDictionaryPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolve home directory: %w", err)
	}

	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(homeDir, "Library", "Application Support", "Dreamtonics", "Voicepeak", "settings", "dic.json"), nil
	case "windows":
		appDataDir := os.Getenv("APPDATA")
		if appDataDir == "" {
			appDataDir = filepath.Join(homeDir, "AppData", "Roaming")
		}
		return filepath.Join(appDataDir, "Dreamtonics", "Voicepeak", "settings", "dic.json"), nil
	default:
		return "", ErrDictionaryPathUnsupported
	}
}

func LoadDefaultDictionary() ([]DictEntry, string, error) {
	path, err := DefaultDictionaryPath()
	if err != nil {
		return nil, "", err
	}

	entries, err := LoadDictionary(path)
	return entries, path, err
}

func LoadDictionary(path string) ([]DictEntry, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []DictEntry{}, nil
		}
		return nil, err
	}

	entries := []DictEntry{}
	if len(data) == 0 {
		return entries, nil
	}

	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("decode dictionary: %w", err)
	}

	return entries, nil
}

func SaveDictionary(path string, entries []DictEntry) error {
	normalized := make([]DictEntry, 0, len(entries))
	for _, entry := range entries {
		entry, err := NormalizeDictEntry(entry)
		if err != nil {
			return err
		}
		normalized = append(normalized, entry)
	}

	sort.SliceStable(normalized, func(i, j int) bool {
		if normalized[i].Surface != normalized[j].Surface {
			return normalized[i].Surface < normalized[j].Surface
		}
		if normalized[i].Pos != normalized[j].Pos {
			return normalized[i].Pos < normalized[j].Pos
		}
		if normalized[i].Pronunciation != normalized[j].Pronunciation {
			return normalized[i].Pronunciation < normalized[j].Pronunciation
		}
		if normalized[i].AccentType != normalized[j].AccentType {
			return normalized[i].AccentType < normalized[j].AccentType
		}
		return normalized[i].Priority < normalized[j].Priority
	})

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create dictionary directory: %w", err)
	}

	tempFile, err := os.CreateTemp(dir, filepath.Base(path)+".tmp-*")
	if err != nil {
		return fmt.Errorf("create temp dictionary file: %w", err)
	}

	tempPath := tempFile.Name()
	defer os.Remove(tempPath)

	encoder := json.NewEncoder(tempFile)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(normalized); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("encode dictionary: %w", err)
	}

	if err := tempFile.Sync(); err != nil {
		_ = tempFile.Close()
		return fmt.Errorf("fsync dictionary temp file: %w", err)
	}

	if err := tempFile.Close(); err != nil {
		return fmt.Errorf("close dictionary temp file: %w", err)
	}

	if err := os.Rename(tempPath, path); err != nil {
		return fmt.Errorf("replace dictionary file: %w", err)
	}

	if dirHandle, err := os.Open(dir); err == nil {
		_ = dirHandle.Sync()
		_ = dirHandle.Close()
	}

	return nil
}

func ExportDictionary(sourcePath, destinationPath string) error {
	entries, err := LoadDictionary(sourcePath)
	if err != nil {
		return err
	}
	return SaveDictionary(destinationPath, entries)
}

func AddDictionaryWord(path string, entry DictEntry) error {
	entry, err := NormalizeDictEntry(entry)
	if err != nil {
		return err
	}

	entries, err := LoadDictionary(path)
	if err != nil {
		return err
	}

	if matchCount := len(FindDictionaryEntriesBySurface(entries, entry.Surface)); matchCount != 0 {
		return fmt.Errorf("%w: surface %q already exists", ErrDictionaryWordConflict, entry.Surface)
	}

	entries = append(entries, entry)
	return SaveDictionary(path, entries)
}

func UpdateDictionaryWordBySurface(path, currentSurface string, nextEntry DictEntry) error {
	currentSurface = normalizeDictionarySurface(currentSurface)
	if currentSurface == "" {
		return fmt.Errorf("%w: current surface is required", ErrDictionaryWordInvalid)
	}

	nextEntry, err := NormalizeDictEntry(nextEntry)
	if err != nil {
		return err
	}

	entries, err := LoadDictionary(path)
	if err != nil {
		return err
	}

	matches := FindDictionaryEntriesBySurface(entries, currentSurface)
	switch len(matches) {
	case 0:
		return fmt.Errorf("%w: surface %q", ErrDictionaryWordNotFound, currentSurface)
	case 1:
	default:
		return fmt.Errorf("%w: surface %q matched %d entries", ErrDictionaryWordConflict, currentSurface, len(matches))
	}

	targetIndex := matches[0]
	for index, entry := range entries {
		if index == targetIndex {
			continue
		}
		if normalizeDictionarySurface(entry.Surface) == nextEntry.Surface {
			return fmt.Errorf("%w: surface %q already exists", ErrDictionaryWordConflict, nextEntry.Surface)
		}
	}

	entries[targetIndex] = nextEntry
	return SaveDictionary(path, entries)
}

func DeleteDictionaryWordBySurface(path, surface string) error {
	surface = normalizeDictionarySurface(surface)
	if surface == "" {
		return fmt.Errorf("%w: surface is required", ErrDictionaryWordInvalid)
	}

	entries, err := LoadDictionary(path)
	if err != nil {
		return err
	}

	matches := FindDictionaryEntriesBySurface(entries, surface)
	switch len(matches) {
	case 0:
		return fmt.Errorf("%w: surface %q", ErrDictionaryWordNotFound, surface)
	case 1:
	default:
		return fmt.Errorf("%w: surface %q matched %d entries", ErrDictionaryWordConflict, surface, len(matches))
	}

	targetIndex := matches[0]
	entries = append(entries[:targetIndex], entries[targetIndex+1:]...)
	return SaveDictionary(path, entries)
}

func ImportDictionary(path string, importedEntries []DictEntry, override bool) error {
	entries, err := LoadDictionary(path)
	if err != nil {
		return err
	}

	for _, importedEntry := range importedEntries {
		importedEntry, err = NormalizeDictEntry(importedEntry)
		if err != nil {
			return err
		}

		matches := FindDictionaryEntriesBySurface(entries, importedEntry.Surface)
		switch len(matches) {
		case 0:
			entries = append(entries, importedEntry)
		case 1:
			if !override {
				return fmt.Errorf("%w: surface %q already exists", ErrDictionaryWordConflict, importedEntry.Surface)
			}
			entries[matches[0]] = importedEntry
		default:
			return fmt.Errorf("%w: surface %q matched %d entries", ErrDictionaryWordConflict, importedEntry.Surface, len(matches))
		}
	}

	return SaveDictionary(path, entries)
}

func FindDictionaryEntriesBySurface(entries []DictEntry, surface string) []int {
	surface = normalizeDictionarySurface(surface)
	indices := []int{}
	for index, entry := range entries {
		if normalizeDictionarySurface(entry.Surface) == surface {
			indices = append(indices, index)
		}
	}
	return indices
}

func NormalizeDictEntry(entry DictEntry) (DictEntry, error) {
	entry.Surface = normalizeDictionarySurface(entry.Surface)
	if entry.Surface == "" {
		return DictEntry{}, fmt.Errorf("%w: surface is required", ErrDictionaryWordInvalid)
	}

	entry.Pronunciation = strings.TrimSpace(entry.Pronunciation)
	if entry.Pronunciation == "" {
		return DictEntry{}, fmt.Errorf("%w: pronunciation is required", ErrDictionaryWordInvalid)
	}
	if !katakanaPattern.MatchString(entry.Pronunciation) {
		return DictEntry{}, fmt.Errorf("%w: pronunciation must be katakana", ErrDictionaryWordInvalid)
	}

	entry.Pos = strings.TrimSpace(entry.Pos)
	if !validDictionaryPos[entry.Pos] {
		return DictEntry{}, fmt.Errorf("%w: pos %q is not supported", ErrDictionaryWordInvalid, entry.Pos)
	}

	if entry.Priority < 0 || entry.Priority > 10 {
		return DictEntry{}, fmt.Errorf("%w: priority must be between 0 and 10", ErrDictionaryWordInvalid)
	}
	if entry.AccentType < 0 {
		return DictEntry{}, fmt.Errorf("%w: accentType must be 0 or greater", ErrDictionaryWordInvalid)
	}

	entry.Lang = strings.TrimSpace(entry.Lang)
	if entry.Lang == "" {
		entry.Lang = defaultDictionaryLang
	}

	return entry, nil
}

func normalizeDictionarySurface(surface string) string {
	surface = strings.Map(func(r rune) rune {
		switch {
		case r == '\u3000':
			return ' '
		case r >= '！' && r <= '～':
			return r - 0xFEE0
		default:
			return r
		}
	}, surface)

	return strings.Join(strings.Fields(strings.TrimSpace(surface)), " ")
}
