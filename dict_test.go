package vpeak

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func sampleDictEntry(surface, pronunciation string) DictEntry {
	return DictEntry{
		Surface:       surface,
		Pronunciation: pronunciation,
		Pos:           "Japanese_Koyuumeishi_ippan",
		Priority:      5,
		AccentType:    0,
		Lang:          "ja",
	}
}

func TestSaveAndLoadDictionary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dic.json")
	entries := []DictEntry{
		sampleDictEntry("GitHub", "ギットハブ"),
		sampleDictEntry("HTTP2", "エイチティーティーピーツー"),
	}

	if err := SaveDictionary(path, entries); err != nil {
		t.Fatalf("SaveDictionary() error = %v", err)
	}

	loaded, err := LoadDictionary(path)
	if err != nil {
		t.Fatalf("LoadDictionary() error = %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("LoadDictionary() count = %d, want 2", len(loaded))
	}
	if loaded[0].Surface != "GitHub" || loaded[1].Surface != "HTTP2" {
		t.Fatalf("LoadDictionary() entries = %+v", loaded)
	}
}

func TestAddDictionaryWordRejectsDuplicateSurface(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dic.json")
	if err := SaveDictionary(path, []DictEntry{sampleDictEntry("GitHub", "ギットハブ")}); err != nil {
		t.Fatalf("SaveDictionary() error = %v", err)
	}

	err := AddDictionaryWord(path, sampleDictEntry("GitHub", "ギットハブ"))
	if !errors.Is(err, ErrDictionaryWordConflict) {
		t.Fatalf("AddDictionaryWord() error = %v, want conflict", err)
	}
}

func TestUpdateDictionaryWordBySurface(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dic.json")
	if err := SaveDictionary(path, []DictEntry{sampleDictEntry("GitHub", "ギットハブ")}); err != nil {
		t.Fatalf("SaveDictionary() error = %v", err)
	}

	nextEntry := sampleDictEntry("GitHub Actions", "ギットハブアクションズ")
	nextEntry.AccentType = 3

	if err := UpdateDictionaryWordBySurface(path, "GitHub", nextEntry); err != nil {
		t.Fatalf("UpdateDictionaryWordBySurface() error = %v", err)
	}

	loaded, err := LoadDictionary(path)
	if err != nil {
		t.Fatalf("LoadDictionary() error = %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("LoadDictionary() count = %d, want 1", len(loaded))
	}
	if loaded[0].Surface != "GitHub Actions" || loaded[0].AccentType != 3 {
		t.Fatalf("updated entry = %+v", loaded[0])
	}
}

func TestDeleteDictionaryWordBySurfaceConflict(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dic.json")
	raw := `[
  {"sur":"git","pron":"ギット","pos":"Japanese_Futsuu_meishi","priority":5,"accentType":0,"lang":"ja"},
  {"sur":"git","pron":"ギット","pos":"Japanese_Futsuu_meishi","priority":5,"accentType":0,"lang":"ja"}
]`
	if err := os.WriteFile(path, []byte(raw), 0o644); err != nil {
		t.Fatalf("os.WriteFile() error = %v", err)
	}

	err := DeleteDictionaryWordBySurface(path, "git")
	if !errors.Is(err, ErrDictionaryWordConflict) {
		t.Fatalf("DeleteDictionaryWordBySurface() error = %v, want conflict", err)
	}
}

func TestImportDictionaryOverride(t *testing.T) {
	path := filepath.Join(t.TempDir(), "dic.json")
	if err := SaveDictionary(path, []DictEntry{sampleDictEntry("GitHub", "ギットハブ")}); err != nil {
		t.Fatalf("SaveDictionary() error = %v", err)
	}

	nextEntry := sampleDictEntry("GitHub", "ギットハブ")
	nextEntry.Pos = "Japanese_Futsuu_meishi"
	nextEntry.AccentType = 1

	if err := ImportDictionary(path, []DictEntry{nextEntry}, true); err != nil {
		t.Fatalf("ImportDictionary() error = %v", err)
	}

	loaded, err := LoadDictionary(path)
	if err != nil {
		t.Fatalf("LoadDictionary() error = %v", err)
	}

	if len(loaded) != 1 {
		t.Fatalf("LoadDictionary() count = %d, want 1", len(loaded))
	}
	if loaded[0].Pos != "Japanese_Futsuu_meishi" || loaded[0].AccentType != 1 {
		t.Fatalf("imported entry = %+v", loaded[0])
	}
}

func TestNormalizeDictEntryDefaultsLangAndNormalizesSurface(t *testing.T) {
	entry, err := NormalizeDictEntry(DictEntry{
		Surface:       " ＧｉｔＨｕｂ　Actions ",
		Pronunciation: "ギットハブアクションズ",
		Pos:           "Japanese_Koyuumeishi_ippan",
		Priority:      5,
		AccentType:    0,
	})
	if err != nil {
		t.Fatalf("NormalizeDictEntry() error = %v", err)
	}

	if entry.Surface != "GitHub Actions" {
		t.Fatalf("NormalizeDictEntry() surface = %q", entry.Surface)
	}
	if entry.Lang != "ja" {
		t.Fatalf("NormalizeDictEntry() lang = %q", entry.Lang)
	}
}
