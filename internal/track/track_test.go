package track

import (
	"strings"
	"testing"
	"time"
)

func TestNewTrack(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name         string
		trackName    string
		trackType    string
		artist       string
		price        float64
		duration     int
		vipMessage   string
		wantTypeName string
		wantErr      bool
	}

	standartCheck := func(t *testing.T, track Track, tc testCase) {
		standartTrack, ok := track.(standard)
		if !ok {
			t.Fatalf("expected type to be %q, got '%T'", "standart", track)
		}

		if standartTrack.artist != tc.artist {
			t.Errorf("expected artist to be %q, got %q", tc.artist, standartTrack.artist)
		}

		if standartTrack.name != tc.trackName {
			t.Errorf("expected name to be %q, got %q", tc.trackName, standartTrack.name)
		}

		if price := int(tc.price * 100); standartTrack.price != price {
			t.Errorf("expected price to be %q, got %q", price, standartTrack.price)
		}

		if duration := time.Duration(tc.duration) * time.Second; standartTrack.duration != duration {
			t.Errorf("expected duration to be %q, got %q", duration, standartTrack.name)
		}

	}

	vipCheck := func(t *testing.T, track Track, tc testCase) {
		vipTrack, ok := track.(vip)
		if !ok {
			t.Fatalf("expected type to be %q, got '%T'", "vip", track)
		}

		if vipTrack.message != tc.vipMessage {
			t.Errorf("expected vip message to be %q, got %q", tc.vipMessage, vipTrack.message)
		}

		standartCheck(t, vipTrack.standard, tc)
	}

	tests := []testCase{
		{
			name:         "standard track success",
			trackType:    "standard",
			trackName:    "A",
			artist:       "B",
			price:        1,
			duration:     1,
			wantTypeName: "standard",
			wantErr:      false,
		},
		{
			name:         "vip track success",
			trackType:    "vip",
			trackName:    "A",
			artist:       "B",
			price:        1,
			duration:     1,
			wantTypeName: "vip",
			wantErr:      false,
		},
		{
			name:      "invalid track type",
			trackType: "bad",
			trackName: "A",
			artist:    "B",
			price:     1,
			duration:  1,
			wantErr:   true,
		},
		{
			name:      "invalid data (missing title/price)",
			trackType: "standard",
			wantErr:   true,
		},

		{
			name:      "invalid data (missing title/price)",
			trackType: "vip",
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			track, err := NewTrack(TrackOpts{
				Type:            tc.trackType,
				Artist:          tc.artist,
				Name:            tc.trackName,
				Price:           tc.price,
				DurationSeconds: tc.duration,
				VIPMessage:      tc.vipMessage,
			})

			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				} else if track != nil {
					t.Fatal("expected nil track on error")
				}
			} else {
				if err != nil {
					t.Fatalf("expected error to be nil, got %v", err)
					return
				}

				switch tc.trackType {
				case "vip":
					vipCheck(t, track, tc)
				case "standard":
					standartCheck(t, track, tc)
				default:
					t.Fatal("invalid test type")
				}
			}
		})
	}
}

func Test_standard_Play(t *testing.T) {
	t.Parallel()

	track := standard{
		duration: 2 * time.Second,
	}

	start := time.Now()
	err := track.Play()
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	min := 2 * time.Second
	max := 3 * time.Second

	if elapsed < min || elapsed > max {
		t.Errorf("expected duration ~2s, got %v", elapsed)
	}
}

func Test_newStandard(t *testing.T) {
	t.Parallel()

	t.Run("returns error when track data is invalid", func(t *testing.T) {
		t.Parallel()

		_, err := newStandard("", "", 0, 1)
		if err == nil {
			t.Fatal("expected error")
		}

		errStr := err.Error()

		expected := []string{
			"track name is missing",
			"track artist is missing",
			"track price has to be bigger than 0.00",
		}

		for _, msg := range expected {
			if !strings.Contains(errStr, msg) {
				t.Errorf("expected error to contain %q, got %v", msg, errStr)
			}
		}
	})

	t.Run("returns standard when track data is valid", func(t *testing.T) {
		t.Parallel()

		track, err := newStandard("Song", "Artist", 1.5, 2)
		if err != nil {
			t.Fatal("unexpected error")
		}

		if track.name != "Song" {
			t.Errorf("expected name %q, got %q", "Song", track.name)
		}

		if track.artist != "Artist" {
			t.Errorf("expected artist %q, got %q", "Artist", track.artist)
		}

		if track.price != 150 {
			t.Errorf("expected price %d, got %d", 150, track.price)
		}

		if track.duration != 2*time.Second {
			t.Errorf("expected duration %v, got %v", 2*time.Second, track.duration)
		}
	})
}

func Test_newVip(t *testing.T) {
	t.Parallel()

	track, err := newVip("Song", "Artist", "msg", 1, 1)
	if err != nil {
		t.Fatal("unexpected error")
	}

	if track.message != "msg" {
		t.Error("wrong message")
	}

	track, err = newVip("", "", "sg", 1, 1)
	if err == nil {
		t.Fatal("expected error")
	}
}
