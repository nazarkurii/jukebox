package jukebox

import (
	"errors"
	"slices"
	"testing"

	"github.com/nazarkurii/jukebox/internal/track"
)

// --- Mocks ---

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
	calcErr    error
}

func (m mockCoinsPolicy) IsValid(amount int) bool {
	return slices.Contains(m.validCoins, amount)
}

func (m mockCoinsPolicy) Denominations() []int {
	return m.validCoins
}

func (m mockCoinsPolicy) CalculateChange(sum int) ([]int, bool, error) {
	return m.change, m.ok, m.calcErr
}

func newMockTrack() mockTrack {
	return mockTrack{title: "A", name: "hello", price: 10}
}

func newTestChosenTrack(totalAccepted int, acceptedDenominations ...int) *chosenTrack {
	mockTrack := newMockTrack()
	return &chosenTrack{
		track:                 mockTrack,
		totalAccepted:         totalAccepted,
		acceptedDenominations: acceptedDenominations,
	}
}

func newTestJukeBox(state State, chochosenTrack *chosenTrack) *JukeBox {
	jb := New([]track.Track{
		newMockTrack(),
	}, mockCoinsPolicy{validCoins: []int{10, 5, 1}})

	jb.chosenTrack = chochosenTrack
	jb.state = state

	return jb
}

func TestNew(t *testing.T) {
	t.Parallel()

	tracks := []track.Track{
		mockTrack{title: "A", price: 10},
	}
	policy := mockCoinsPolicy{validCoins: []int{10}}
	jb := New(tracks, policy)

	if jb.State() != StateIdle {
		t.Error("expected idle state")
	}

	if len(jb.tracks) != len(tracks) {
		t.Fatalf("expected %d tracks, got %d", len(tracks), len(jb.tracks))
	}

	expectedInfo := TrackInfo{Price: 10, Title: "A", Number: 1}
	if jb.trackList[0] != expectedInfo {
		t.Errorf("expected track info %v, got %v", expectedInfo, jb.trackList[0])
	}
}

func TestJukeBox_ChooseTrack(t *testing.T) {
	t.Parallel()

	runMethods := func(t *testing.T, name string, choose func(jb *JukeBox) (string, int, error)) {
		t.Run(name, func(t *testing.T) {

			t.Run("successfuly changes state and sets chosen track", func(t *testing.T) {
				jb := newTestJukeBox(StateIdle, nil)

				title, price, err := choose(jb)

				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if title != "A" || price != 10 {
					t.Errorf("expected title 'A' and price 10, got %q and %d", title, price)
				}
				if jb.State() != StateAcceptCoins {
					t.Errorf("expected state to change to AcceptCoins, got %v", jb.State())
				}
				if jb.chosenTrack == nil {
					t.Error("expected chosenTrack to be set")
				}
			})

			t.Run("returns ErrTrackNotFound when track does not exist", func(t *testing.T) {
				jb := New(nil, mockCoinsPolicy{})

				title, price, err := choose(jb)

				if !errors.Is(err, ErrTrackNotFound) {
					t.Fatalf("expected error %v, got %v", ErrTrackNotFound, err)
				}
				if title != "" || price != 0 {
					t.Error("expected zero values for title and price when error")
				}
				if jb.State() != StateIdle {
					t.Errorf("expected state to stay Idle, got %v", jb.State())
				}
			})

			t.Run("returns ErrInvalidOperation when state is invalid", func(t *testing.T) {
				jb := newTestJukeBox(StateAcceptCoins, newTestChosenTrack(0))

				title, price, err := choose(jb)

				if !errors.Is(err, ErrInvalidOperation) {
					t.Fatalf("expected error %v, got %v", ErrInvalidOperation, err)
				}
				if title != "" || price != 0 {
					t.Error("expected zero values for title and price when error")
				}
				if jb.State() != StateAcceptCoins {
					t.Errorf("expected state to stay AcceptCoins, got %v", jb.State())
				}
			})
		})
	}

	runMethods(t, "ChooseTrackByNumber", func(jb *JukeBox) (string, int, error) {
		return jb.ChooseTrackByNumber(1)
	})

	runMethods(t, "ChooseTrackByName", func(jb *JukeBox) (string, int, error) {
		return jb.ChooseTrackByName("hello")
	})
}

func TestJukeBox_AcceptCoin(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                  string
		setup                 func() *JukeBox
		coin, wantTotal       int
		wantState, setupState State

		wantErr error
	}{
		{
			name:       "returns error when state is invalid",
			setupState: StateIdle,
			coin:       5,
			wantTotal:  0,
			wantState:  StateIdle,
			wantErr:    ErrInvalidOperation,
		},
		{
			name:       "returns error when invalid denomination",
			setupState: StateAcceptCoins,
			coin:       7,
			wantErr:    ErrInvalidCoinDenomination,
		},
		{
			name:       "returns total and price when 'coin + total < price'",
			setupState: StateAcceptCoins,
			coin:       5,
			wantTotal:  5,
			wantState:  StateAcceptCoins,
		},
		{
			name:       "returns total, price and changes state when 'coin + total >= price'",
			setupState: StateAcceptCoins,
			coin:       10,
			wantTotal:  10,
			wantState:  StatePlaying,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jb := newTestJukeBox(tc.setupState, newTestChosenTrack(0))
			total, _, err := jb.AcceptCoin(tc.coin)

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected error %v, got %v", tc.wantErr, err)
			}
			if err == nil {
				if total != tc.wantTotal {
					t.Errorf("expected total %d, got %d", tc.wantTotal, total)
				}
				if jb.State() != tc.wantState {
					t.Errorf("expected state %v, got %v", tc.wantState, jb.State())
				}
			}
		})
	}
}

func TestJukeBox_CancelTrack(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		setupState      State
		setupTrack      *chosenTrack
		wantCoins       []int
		wantState       State
		wantChosenTrack *chosenTrack
	}{
		{
			name:            "returns accepted coins and resets state when in StateAcceptCoins",
			setupState:      StateAcceptCoins,
			setupTrack:      newTestChosenTrack(15, 10, 5),
			wantCoins:       []int{10, 5},
			wantState:       StateIdle,
			wantChosenTrack: nil,
		},
		{
			name:            "returns nil and does not change state when in StateIdle",
			setupState:      StateIdle,
			setupTrack:      nil,
			wantCoins:       nil,
			wantState:       StateIdle,
			wantChosenTrack: nil,
		},
		{
			name:            "returns nil and does not change state when in StatePlaying",
			setupState:      StatePlaying,
			setupTrack:      newTestChosenTrack(10, 10),
			wantCoins:       nil,
			wantState:       StatePlaying,
			wantChosenTrack: newTestChosenTrack(10, 10),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			jb := newTestJukeBox(tc.setupState, tc.setupTrack)

			coins := jb.CancelTrack()

			if !slices.Equal(coins, tc.wantCoins) {
				t.Errorf("expected coins %v, got %v", tc.wantCoins, coins)
			}

			if jb.State() != tc.wantState {
				t.Errorf("expected state %v, got %v", tc.wantState, jb.State())
			}

			if tc.wantChosenTrack == nil {
				if jb.chosenTrack != nil {
					t.Error("expected chosenTrack to be nil")
				}
			} else {
				if jb.chosenTrack == nil {
					t.Error("expected chosenTrack not to be changed, but got nil")
				} else if jb.chosenTrack.totalAccepted != tc.wantChosenTrack.totalAccepted {
					t.Errorf("expected the same track with %d accepted, got %d",
						tc.wantChosenTrack.totalAccepted, jb.chosenTrack.totalAccepted)
				}
			}
		})
	}
}
func TestJukeBox_GetChosenTrackTitle(t *testing.T) {
	t.Parallel()

	ct := newTestChosenTrack(0)

	tests := []struct {
		name        string
		sate        State
		wantErr     error
		wantTitle   string
		chosenTrack *chosenTrack
	}{
		{
			name:      "returns error when state is invalid",
			sate:      StateIdle,
			wantErr:   ErrInvalidOperation,
			wantTitle: "",
		},
		{
			name:        "returns title when state is 'StateAcceptCoins",
			sate:        StateAcceptCoins,
			wantErr:     nil,
			wantTitle:   ct.track.Title(),
			chosenTrack: ct,
		},
		{
			name:        "returns title when state is 'StatePlaying",
			sate:        StatePlaying,
			wantErr:     nil,
			wantTitle:   ct.track.Title(),
			chosenTrack: ct,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			jb := newTestJukeBox(tc.sate, tc.chosenTrack)
			title, err := jb.GetChosenTrackTitle()
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected error to be '%s', got '%s'", tc.wantErr, err)
			} else if title != tc.wantTitle {
				t.Errorf("expected title to be '%s', got '%s'", tc.wantTitle, title)
			}
		})
	}

}

func TestJukeBox_PlayChosenTrack(t *testing.T) {
	t.Parallel()

	customErr := errors.New("mechanical failure")

	tests := []struct {
		name        string
		setup       func() *JukeBox
		wantChange  []int
		wantHistory []string
		wantErr     error
	}{
		{
			name: "returns error when state is invalid",
			setup: func() *JukeBox {
				jb := newTestJukeBox(StateIdle, nil)
				return jb
			},
			wantErr: ErrInvalidOperation,
		},

		{
			name: "returns error along with acceptedMoney when fails to play the track",
			setup: func() *JukeBox {
				chosenTrack := newTestChosenTrack(5, 5)
				mockTrack := newMockTrack()
				mockTrack.playErr = customErr
				chosenTrack.track = mockTrack
				jb := newTestJukeBox(StatePlaying, chosenTrack)
				return jb
			},
			wantChange: []int{5},
			wantErr:    customErr,
		},

		{
			name: "returns err when fails to calculate the change",
			setup: func() *JukeBox {
				jb := newTestJukeBox(StatePlaying, newTestChosenTrack(5, 5))
				jb.coins = mockCoinsPolicy{calcErr: customErr}
				return jb
			},
			wantHistory: []string{newMockTrack().title},
			wantErr:     customErr,
		},
		{
			name: "returns err when imposible to calculate the change",
			setup: func() *JukeBox {
				jb := newTestJukeBox(StatePlaying, newTestChosenTrack(5, 5))
				jb.coins = mockCoinsPolicy{ok: false}
				return jb
			},
			wantErr:     ErrImpossibleChange,
			wantHistory: []string{newMockTrack().title},
		},

		{
			name: "returns change when total accepted is greater than the price",
			setup: func() *JukeBox {
				jb := newTestJukeBox(StatePlaying, newTestChosenTrack(10, 5))
				jb.coins = mockCoinsPolicy{ok: true, change: []int{5}}
				return jb
			},
			wantHistory: []string{newMockTrack().title},
			wantChange:  []int{5},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			jb := tc.setup()
			change, err := jb.PlayChosenTrack()

			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("expected error containing %s, got %s", tc.wantErr, err)
			}

			if !slices.Equal(change, tc.wantChange) {
				t.Errorf("expected change %v, got %v", tc.wantChange, change)
			}

			if jb.State() != StateIdle {
				t.Error("has to reset to idle after any play result")
			}
		})
	}
}
