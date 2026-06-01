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
