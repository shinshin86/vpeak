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
