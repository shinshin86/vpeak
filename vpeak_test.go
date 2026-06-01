package vpeak

import "testing"

func TestParseEmotion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Emotion
		wantErr bool
	}{
		{"empty string", "", Emotion{}, false},
		{"single emotion without value", "happy", Emotion{Happy: 100}, false},
		{"single emotion with value", "happy=50", Emotion{Happy: 50}, false},
		{"multiple emotions", "happy=40,fun=60", Emotion{Happy: 40, Fun: 60}, false},
		{"all four emotions", "happy=10,fun=20,angry=30,sad=40", Emotion{Happy: 10, Fun: 20, Angry: 30, Sad: 40}, false},
		{"boundary zero", "happy=0", Emotion{Happy: 0}, false},
		{"boundary max", "happy=100", Emotion{Happy: 100}, false},
		{"duplicate keys last wins", "happy=30,happy=70", Emotion{Happy: 70}, false},
		{"invalid emotion name", "joyful=50", Emotion{}, true},
		{"non-numeric value", "happy=abc", Emotion{}, true},
		{"negative value", "happy=-1", Emotion{}, true},
		{"value over max", "happy=101", Emotion{}, true},
		{"empty value", "happy=", Emotion{}, true},
		{"trailing comma", "happy=50,", Emotion{}, true},
		{"leading comma", ",happy=50", Emotion{}, true},
		{"empty middle segment", "happy=50,,sad=50", Emotion{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseEmotion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseEmotion(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseEmotion(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

func TestEmotionString(t *testing.T) {
	tests := []struct {
		name string
		e    Emotion
		want string
	}{
		{"zero value", Emotion{}, ""},
		{"only happy", Emotion{Happy: 50}, "happy=50"},
		{"happy and fun", Emotion{Happy: 40, Fun: 60}, "happy=40,fun=60"},
		{"all four in fixed order", Emotion{Happy: 10, Sad: 40, Angry: 30, Fun: 20}, "happy=10,fun=20,angry=30,sad=40"},
		{"zero values are omitted", Emotion{Happy: 0, Sad: 50}, "sad=50"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.String(); got != tt.want {
				t.Errorf("Emotion.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestEmotionIsZero(t *testing.T) {
	tests := []struct {
		name string
		e    Emotion
		want bool
	}{
		{"all zero", Emotion{}, true},
		{"happy set", Emotion{Happy: 1}, false},
		{"sad set", Emotion{Sad: 1}, false},
		{"angry set", Emotion{Angry: 1}, false},
		{"fun set", Emotion{Fun: 1}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.IsZero(); got != tt.want {
				t.Errorf("Emotion.IsZero() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateEmotionExpression(t *testing.T) {
	allowed := []string{"amaama", "aori", "live"}
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"empty", "", "", false},
		{"single name", "amaama", "amaama", false},
		{"single weighted", "amaama=50", "amaama=50", false},
		{"multiple weighted", "amaama=40,live=60", "amaama=40,live=60", false},
		{"trim", " amaama=40,live=60 ", "amaama=40,live=60", false},
		{"zero", "amaama=0", "amaama=0", false},
		{"max", "amaama=100", "amaama=100", false},
		{"unknown name", "happy=50", "", true},
		{"empty value", "amaama=", "", true},
		{"empty segment", "amaama=50,,live=50", "", true},
		{"leading comma", ",amaama=50", "", true},
		{"trailing comma", "amaama=50,", "", true},
		{"non integer", "amaama=foo", "", true},
		{"negative", "amaama=-1", "", true},
		{"too high", "amaama=101", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ValidateEmotionExpression(tt.input, allowed)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateEmotionExpression(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("ValidateEmotionExpression(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestResolveNarratorName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"legacy alias", "f1", "Japanese Female 1"},
		{"formal name", "Zundamon", "Zundamon"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := resolveNarratorName(tt.input); got != tt.want {
				t.Fatalf("resolveNarratorName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseVoicepeakListOutput(t *testing.T) {
	output := `[debug][1780286975][voicepeak.GeneralDebug] UserApplication Folder: /Users/example/Library/Application Support/Dreamtonics/Voicepeak
iconv_open is not supported
Tohoku Zunko
Zundamon

`

	got := parseVoicepeakListOutput(output)
	want := []string{"Tohoku Zunko", "Zundamon"}
	if len(got) != len(want) {
		t.Fatalf("parseVoicepeakListOutput() = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("parseVoicepeakListOutput()[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestNormalizeEmotionExpression(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"empty", "", "", false},
		{"bare name expands to 100", "happy", "happy=100", false},
		{"dynamic bare name expands to 100", "amaama", "amaama=100", false},
		{"explicit value kept", "happy=50", "happy=50", false},
		{"multiple", "happy=40,fun=60", "happy=40,fun=60", false},
		{"dynamic names kept", "amaama=40,live=60", "amaama=40,live=60", false},
		{"input order preserved", "sad=10,happy=20", "sad=10,happy=20", false},
		{"trim outer whitespace", " amaama=40,live=60 ", "amaama=40,live=60", false},
		{"zero dropped", "happy=0", "", false},
		{"zero dropped among others", "happy=0,fun=60", "fun=60", false},
		{"max", "happy=100", "happy=100", false},
		{"empty value", "happy=", "", true},
		{"empty segment", "happy=50,,fun=50", "", true},
		{"leading comma", ",happy=50", "", true},
		{"trailing comma", "happy=50,", "", true},
		{"non integer", "happy=foo", "", true},
		{"negative", "happy=-1", "", true},
		{"too high", "happy=101", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeEmotionExpression(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("normalizeEmotionExpression(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.want {
				t.Fatalf("normalizeEmotionExpression(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildOptionsEmotion(t *testing.T) {
	flagValue := func(opts []string, flag string) (string, bool) {
		for i, o := range opts {
			if o == flag && i+1 < len(opts) {
				return opts[i+1], true
			}
		}
		return "", false
	}

	t.Run("bare emotion expands to 100", func(t *testing.T) {
		opts := buildOptions("hi", Options{Narrator: "f1", Emotion: "happy"})
		if v, _ := flagValue(opts, "--narrator"); v != "Japanese Female 1" {
			t.Fatalf("--narrator = %q, want %q", v, "Japanese Female 1")
		}
		if v, _ := flagValue(opts, "--emotion"); v != "happy=100" {
			t.Fatalf("--emotion = %q, want %q", v, "happy=100")
		}
	})

	t.Run("dynamic emotion names pass through", func(t *testing.T) {
		opts := buildOptions("hi", Options{Narrator: "Zundamon", Emotion: "amaama=40,live=60"})
		if v, _ := flagValue(opts, "--narrator"); v != "Zundamon" {
			t.Fatalf("--narrator = %q, want %q", v, "Zundamon")
		}
		if v, _ := flagValue(opts, "--emotion"); v != "amaama=40,live=60" {
			t.Fatalf("--emotion = %q, want %q", v, "amaama=40,live=60")
		}
	})

	t.Run("all-zero emotion omits --emotion", func(t *testing.T) {
		opts := buildOptions("hi", Options{Narrator: "f1", Emotion: "happy=0"})
		if _, ok := flagValue(opts, "--emotion"); ok {
			t.Fatalf("--emotion should be omitted for all-zero emotion, got %v", opts)
		}
	})

	t.Run("no emotion omits --emotion", func(t *testing.T) {
		opts := buildOptions("hi", Options{Narrator: "f1"})
		if _, ok := flagValue(opts, "--emotion"); ok {
			t.Fatalf("--emotion should be omitted when no emotion given, got %v", opts)
		}
	})
}
