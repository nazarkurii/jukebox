package jukebox

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/nazarkurii/jukebox/internal/track"
)

type mockTrack struct {
	title   string
	name    string
	price   int
	playErr error
}

func (m mockTrack) Title() string        { return m.title }
func (m mockTrack) TitleAligned() string { return m.title }
func (m mockTrack) Name() string         { return m.name }
func (m mockTrack) Price() int           { return m.price }
func (m mockTrack) Play() error          { return m.playErr }

type mockCoinsPolicy struct {
	validCoins []int
	change     []int
	ok         bool
}

func testJuxBox() *JukeBox {
	return New([]track.Track{
		mockTrack{title: "A", name: "hello", price: 10},
	}, mockCoinsPolicy{})
}

func (m mockCoinsPolicy) IsValid(amount int) bool {
	return slices.Contains(m.validCoins, amount)
}

func (m mockCoinsPolicy) Denominations() []int {
	return m.validCoins
}

func (m mockCoinsPolicy) CalculateChange(sum int) ([]int, bool) {
	return m.change, m.ok
}

func TestNew(t *testing.T) {
	t.Parallel()

	tracks := []track.Track{
		mockTrack{title: "A", price: 10},
	}

	jb := New(tracks, mockCoinsPolicy{})

	if jb.State() != StateIdle {
		t.Error("expected idle state")
	}

	if !slices.Equal(jb.tracks, tracks) {
		t.Errorf("expected tracks to be %v, got %v", tracks, jb.tracks)
	}

	if !slices.Equal(jb.trackList, []TrackInfo{
		{Price: 10, Title: "A", Number: 1},
	}) {
		t.Errorf("expected tracks to be %v, got %v", tracks, jb.tracks)
	}

	if _, ok := jb.coins.(mockCoinsPolicy); !ok {
		t.Error("expected coins to be of 'mockCoinsPolicy' type")
	}

	if jb.chosenTrack != nil {
		t.Error("expexted chosen track to be nil")
	}

	if jb.history != nil {
		t.Error("expexted history to be nil")
	}

}

func TestJukeBox_TrackList(t *testing.T) {
	t.Parallel()

	jb := testJuxBox()

	list := jb.TrackList()

	if len(list) != 1 {
		t.Fatal("expected 1 track")
	}

	if list[0].Title != "A" {
		t.Error("wrong title")
	}

	if list[0].Number != 1 {
		t.Error("wrong number")
	}

	if list[0].Price != 10 {
		t.Error("wrong price")
	}

	list[0].Title = "B"
	list = jb.TrackList()

	if list[0].Title != "A" {
		t.Error("expected a copy of a slice")
	}
}

func TestJukeBox_ChooseTrack(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		methodFound    func(jb *JukeBox) (string, int, error)
		methodNotFound func(jb *JukeBox) (string, int, error)
	}{
		{
			name: "by number",
			methodFound: func(jb *JukeBox) (string, int, error) {
				return jb.ChooseTrackByNumber(1)
			},
			methodNotFound: func(jb *JukeBox) (string, int, error) {
				return jb.ChooseTrackByNumber(2)
			},
		},

		{
			name: "by name",
			methodFound: func(jb *JukeBox) (string, int, error) {
				return jb.ChooseTrackByName("hello")
			},
			methodNotFound: func(jb *JukeBox) (string, int, error) {
				return jb.ChooseTrackByName("wrong")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			t.Run("found", func(t *testing.T) {
				jb := testJuxBox()

				title, price, err := tc.methodFound(jb)

				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}

				if title != "A" || price != 10 {
					t.Error("wrong track data")
				}

				if jb.state != StateAcceptCoins {
					t.Error("wrong state")
				}
			})

			t.Run("not found", func(t *testing.T) {
				jb := testJuxBox()

				title, price, err := tc.methodNotFound(jb)

				if !errors.Is(err, ErrTrackNotFound) {
					t.Fatalf("expect *testing.T,ted error %v, got %v", ErrTrackNotFound, err)
				}

				if title != "" || price != 0 {
					t.Fatal("expected title and price to be zero values")
				}
			})

			t.Run("invalid operation", func(t *testing.T) {
				jb := testJuxBox()
				jb.state = StateAcceptCoins

				title, price, err := tc.methodFound(jb)

				if !errors.Is(err, ErrInvalidOperation) {
					t.Fatalf("expect *testing.T,ted error %v, got %v", ErrTrackNotFound, err)
				}

				if title != "" || price != 0 {
					t.Fatal("expected title and price to be zero values")
				}
			})
		})
	}

}

func TestJukeBox_AcceptCoin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		setup     func() *JukeBox
		coin      int
		wantTotal int
		wantPrice int
		wantState State
		wantErr   error
	}{
		{
			name: "invalid state",
			setup: func() *JukeBox {
				return &JukeBox{
					state: StateIdle,
					coins: mockCoinsPolicy{validCoins: []int{10}},
				}
			},
			coin:      10,
			wantErr:   ErrInvalidOperation,
			wantState: StateIdle,
			wantTotal: 0,
			wantPrice: 0,
		},
		{
			name: "invalid coin denomination",
			setup: func() *JukeBox {
				return &JukeBox{
					state: StateAcceptCoins,
					coins: mockCoinsPolicy{validCoins: []int{5}},
					chosenTrack: &chosenTrack{
						track: mockTrack{price: 10},
					},
				}
			},
			coin:      10,
			wantErr:   ErrInvalidCoinDenomination,
			wantState: StateAcceptCoins,
			wantTotal: 0,
			wantPrice: 0,
		},
		{
			name: "valid coin but not enough",
			setup: func() *JukeBox {
				return &JukeBox{
					state: StateAcceptCoins,
					coins: mockCoinsPolicy{validCoins: []int{5}},
					chosenTrack: &chosenTrack{
						track:         mockTrack{price: 10},
						totalAccepted: 0,
					},
				}
			},
			coin:      5,
			wantTotal: 5,
			wantPrice: 10,
			wantState: StateAcceptCoins,
		},
		{
			name: "valid coin and enough to play",
			setup: func() *JukeBox {
				return &JukeBox{
					state: StateAcceptCoins,
					coins: mockCoinsPolicy{validCoins: []int{10}},
					chosenTrack: &chosenTrack{
						track:         mockTrack{price: 10},
						totalAccepted: 0,
					},
				}
			},
			coin:      10,
			wantTotal: 10,
			wantPrice: 10,
			wantState: StatePlaying,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			jb := tc.setup()
			total, price, err := jb.AcceptCoin(tc.coin)

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if total != tc.wantTotal || price != tc.wantPrice {
				t.Errorf("wrong values: got total=%d price=%d, want total=%d price=%d",
					total, price, tc.wantTotal, tc.wantPrice)
			}

			if jb.state != tc.wantState {
				t.Errorf("wrong state: got %v want %v", jb.state, tc.wantState)
			}
		})
	}
}

func TestJukeBox_CancelTrack(t *testing.T) {
	t.Parallel()

	t.Run("cancel returns coins", func(t *testing.T) {
		t.Parallel()

		jb := testJuxBox()

		jb.state = StateAcceptCoins
		jb.chosenTrack = &chosenTrack{
			track:                 mockTrack{},
			totalAccepted:         200,
			acceptedDenominations: []int{100, 50, 50},
		}

		coins := jb.CancelTrack()

		if !slices.Equal(coins, []int{100, 50, 50}) {
			t.Error("wrong coins returned")
		}

		if jb.State() != StateIdle {
			t.Error("should reset state")
		}
	})

	t.Run("cancel in wrong state", func(t *testing.T) {
		t.Parallel()

		jb := New(nil, mockCoinsPolicy{})

		coins := jb.CancelTrack()

		if coins != nil {
			t.Error("expected nil")
		}
	})
}

func TestJukeBox_PlayChosenTrack(t *testing.T) {
	t.Parallel()

	playErr := errors.New("play error")

	tests := []struct {
		name        string
		setup       func() *JukeBox
		wantTitle   string
		wantErr     error
		wantPlayErr error
		wantChange  []int
	}{
		{
			name: "invalid state",
			setup: func() *JukeBox {
				return testJuxBox()
			},
			wantErr: ErrInvalidOperation,
		},
		{
			name: "track play error returns inserted coins",
			setup: func() *JukeBox {
				jb := testJuxBox()
				jb.state = StatePlaying

				tr := &mockTrack{
					title:   "title_example",
					price:   100,
					playErr: playErr,
				}

				jb.chosenTrack = &chosenTrack{
					track:                 tr,
					acceptedDenominations: []int{100},
					totalAccepted:         100,
				}

				return jb
			},
			wantTitle:   "title_example",
			wantPlayErr: fmt.Errorf("failed to play the track: %w", playErr),
			wantChange:  []int{100},
		},
		{
			name: "impossible change",
			setup: func() *JukeBox {
				jb := testJuxBox()
				jb.state = StatePlaying

				tr := &mockTrack{
					title: "title_example",
					price: 100,
				}

				jb.coins = mockCoinsPolicy{
					ok: false,
				}

				jb.chosenTrack = &chosenTrack{
					track:                 tr,
					acceptedDenominations: []int{200},
					totalAccepted:         200,
				}

				return jb
			},
			wantTitle:   "title_example",
			wantPlayErr: ErrImpossibleChange,
		},
		{
			name: "no change",
			setup: func() *JukeBox {
				jb := testJuxBox()
				jb.state = StatePlaying

				tr := &mockTrack{
					title: "title_example",
					price: 100,
				}

				jb.coins = mockCoinsPolicy{
					ok:     true,
					change: []int{},
				}

				jb.chosenTrack = &chosenTrack{
					track:                 tr,
					acceptedDenominations: []int{100},
					totalAccepted:         100,
				}

				return jb
			},
			wantTitle:  "title_example",
			wantChange: []int{},
		},
		{
			name: "with change",
			setup: func() *JukeBox {
				jb := testJuxBox()
				jb.state = StatePlaying

				tr := &mockTrack{
					title: "title_example",
					price: 100,
				}

				jb.coins = mockCoinsPolicy{
					ok:     true,
					change: []int{50, 50},
				}

				jb.chosenTrack = &chosenTrack{
					track:                 tr,
					acceptedDenominations: []int{200},
					totalAccepted:         200,
				}

				return jb
			},
			wantTitle:  "title_example",
			wantChange: []int{50, 50},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			title, playFn, err := tc.setup().PlayChosenTrack()

			if tc.wantErr != nil {
				if !errors.Is(err, tc.wantErr) {
					t.Fatalf("expected error %v, got %v", tc.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if title != tc.wantTitle {
				t.Errorf("expected title %q, got %q", tc.wantTitle, title)
			}

			change, playErr := playFn()

			if tc.wantPlayErr != nil {
				if !errors.Is(playErr, tc.wantPlayErr) &&
					!strings.Contains(playErr.Error(), tc.wantPlayErr.Error()) {
					t.Fatalf("expected play error %v, got %v", tc.wantPlayErr, playErr)
				}
				return
			}

			if playErr != nil {
				t.Fatalf("unexpected play error: %v", playErr)
			}

			if !slices.Equal(change, tc.wantChange) {
				t.Errorf("expected change %v, got %v", tc.wantChange, change)
			}
		})
	}
}
