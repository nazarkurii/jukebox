package track

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type Track interface {
	Title() string
	TitleAligned() string
	Name() string
	Price() int
	Play() error
}

func NewTrack(trackType, artist, name, vipMessage string, price float64, durationSeconds int) (Track, error) {
	var err error
	var track Track

	switch trackType {
	case "standard":
		track, err = newStandard(name, artist, price, durationSeconds)
	case "vip":
		track, err = newVip(name, artist, vipMessage, price, durationSeconds)
	default:
		err = fmt.Errorf("invalid track type, hast to be either 'standard' or 'vip', got '%s'", trackType)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create '%s' track: %w", trackType, err)
	}

	return track, err
}

type standard struct {
	name, artist string
	price        int
	duration     time.Duration
}

func (b standard) Title() string {
	return b.artist + " - " + b.name
}

func (b standard) TitleAligned() string {
	return fmt.Sprintf("%-23s - %s", b.artist, b.name)
}

func (b standard) Artist() string {
	return b.artist
}

func (b standard) Name() string {
	return b.name
}

func (b standard) Price() int {
	return b.price
}

func (b standard) Play() error {
	const width = 30

	ticker := time.NewTicker(time.Millisecond * 200)
	defer ticker.Stop()

	var elapsed time.Duration
	for range ticker.C {
		elapsed += time.Millisecond * 200

		progress := float64(elapsed) / float64(b.duration)
		filled := int(progress * width)

		fmt.Printf(
			"\r|[%s>%s]| %5.1f%%",
			strings.Repeat("=", filled),
			strings.Repeat("-", width-filled),
			progress*100,
		)

		if elapsed == b.duration {
			break
		}
	}

	return nil
}

type vip struct {
	standard
	message string
}

func (v vip) TitleAligned() string {
	return "[VIP] " + v.standard.TitleAligned()
}

func (v vip) Title() string {
	return "[VIP] " + v.standard.Title()
}

func (v vip) Play() error {
	if v.message != "" {
		fmt.Println(v.message)
	}

	return v.standard.Play()
}

func newStandard(name, artist string, price float64, durationSeconds int) (standard, error) {
	var errs error

	if name == "" {
		errs = errors.New("track name is missing")
	}

	if artist == "" {
		errs = errors.Join(errs, errors.New("track artist is missing"))
	}

	if price == 0 {
		errs = errors.Join(errs, errors.New("track price has to be bigger than 0.00"))
	}

	if errs != nil {
		return standard{}, errs
	}

	return standard{
		name:     name,
		artist:   artist,
		price:    int(price * 100),
		duration: time.Duration(durationSeconds) * time.Second,
	}, nil
}

func newVip(name, artist, msg string, price float64, durationSeconds int) (vip, error) {
	standard, err := newStandard(name, artist, price, durationSeconds)
	if err != nil {
		return vip{}, err
	}

	return vip{
		standard: standard,
		message:  msg,
	}, nil
}
