package jukebox

import (
	"errors"
	"fmt"
	"strings"

	"github.com/nazarkurii/jukebox/internal/track"
)

type State int

const (
	StateIdle State = iota
	StateAcceptCoins
	StatePlaying
)

var (
	ErrTrackNotFound           = errors.New("non-existing track")
	ErrInvalidOperation        = errors.New("invalid operation")
	ErrImpossibleChange        = errors.New("impossible change")
	ErrInvalidCoinDenomination = errors.New("invalid coin denimination")
)

type CoinsPolicy interface {
	IsValid(denomination int) bool
	Denominations() []int
	CalculateChange(changeSum int) ([]int, bool)
}

type TrackInfo struct {
	Title  string
	Number int
	Price  int
}

type chosenTrack struct {
	track                 track.Track
	acceptedDenominations []int
	totalAccepted         int
}

type JukeBox struct {
	state       State
	tracks      []track.Track
	trackList   []TrackInfo
	coins       CoinsPolicy
	chosenTrack *chosenTrack
	history     []string
}

func New(tracks []track.Track, coins CoinsPolicy) *JukeBox {
	jb := &JukeBox{
		state:     StateIdle,
		trackList: make([]TrackInfo, len(tracks)),
		tracks:    make([]track.Track, len(tracks)),
		coins:     coins,
	}

	for i, track := range tracks {
		jb.trackList[i] = TrackInfo{
			Price:  track.Price(),
			Title:  track.Title(),
			Number: i + 1,
		}

		jb.tracks[i] = track
	}

	return jb
}

func (jb *JukeBox) changeState(state State) {
	jb.state = state
}

func (jb *JukeBox) State() State {
	return jb.state
}

func (jb *JukeBox) History() []string {
	return jb.history
}

func (jb *JukeBox) TrackList() []TrackInfo {
	trackList := make([]TrackInfo, len(jb.trackList))
	copy(trackList, jb.trackList)
	return trackList
}

func (jb *JukeBox) CoinDenominations() []int {
	return jb.coins.Denominations()
}

func (jb *JukeBox) chooseTrack(compare func(number int, track track.Track) bool) (string, int, error) {
	if jb.state != StateIdle {
		return "", 0, ErrInvalidOperation
	}

	for i, track := range jb.tracks {
		if compare(i+1, track) {
			jb.changeState(StateAcceptCoins)
			jb.chosenTrack = &chosenTrack{
				track: track,
			}
			return track.Title(), track.Price(), nil
		}
	}

	return "", 0, ErrTrackNotFound
}

func (jb *JukeBox) ChooseTrackByNumber(n int) (string, int, error) {
	return jb.chooseTrack(func(number int, track track.Track) bool {
		return n == number
	})
}

func (jb *JukeBox) ChooseTrackByName(name string) (string, int, error) {
	return jb.chooseTrack(func(number int, track track.Track) bool {
		return strings.HasPrefix(strings.ToLower(track.Name()), strings.ToLower(name))
	})
}

func (jb *JukeBox) AcceptCoin(denomination int) (int, int, error) {
	if jb.state != StateAcceptCoins {
		return 0, 0, ErrInvalidOperation
	}

	if !jb.coins.IsValid(denomination) {
		return 0, 0, ErrInvalidCoinDenomination
	}

	jb.chosenTrack.acceptedDenominations = append(jb.chosenTrack.acceptedDenominations, denomination)
	jb.chosenTrack.totalAccepted += denomination

	if jb.chosenTrack.totalAccepted >= jb.chosenTrack.track.Price() {
		jb.changeState(StatePlaying)
	}

	return jb.chosenTrack.totalAccepted, jb.chosenTrack.track.Price(), nil
}

func (jb *JukeBox) CancelTrack() []int {
	if jb.state != StateAcceptCoins {
		return nil
	}

	acceptedDenominations := jb.chosenTrack.acceptedDenominations
	jb.chosenTrack = nil

	jb.changeState(StateIdle)
	return acceptedDenominations
}

func (jb *JukeBox) PlayChosenTrack() (string, func() ([]int, error), error) {
	if jb.state != StatePlaying {
		return "", nil, ErrInvalidOperation
	}

	return jb.chosenTrack.track.Title(), func() ([]int, error) {
		defer jb.changeState(StateIdle)

		playedTrack := jb.chosenTrack
		jb.chosenTrack = nil
		jb.history = append(jb.history, playedTrack.track.TitleAligned())

		err := playedTrack.track.Play()
		if err != nil {
			return playedTrack.acceptedDenominations, fmt.Errorf("failed to play the track: %w", err)
		}

		change, ok := jb.coins.CalculateChange(playedTrack.totalAccepted - playedTrack.track.Price())
		if !ok {
			return nil, fmt.Errorf("failed to calculate the change: %w", ErrImpossibleChange)
		}

		return change, nil
	}, nil
}
