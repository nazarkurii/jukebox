package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/nazarkurii/jukebox/internal/track"
)

type Track struct {
	Type       string  `json:"type"`
	Artist     string  `json:"artist"`
	Title      string  `json:"title"`
	VipMessage string  `json:"vip_message"`
	Duration   int     `json:"duration"`
	Price      float64 `json:"price"`
}

func LoadTracksFromJSON(path string) ([]track.Track, error) {
	if path == "" {
		return nil, errors.New("provided file path is blank")
	}

	fileBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}

	var config struct {
		Tracks []Track `json:"tracks"`
	}

	err = json.Unmarshal(fileBytes, &config)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal content of config file: %w", err)
	}

	var tracks []track.Track
	var errs error

	for _, trackCfg := range config.Tracks {
		track, err := track.NewTrack(trackCfg.Type, trackCfg.Artist, trackCfg.Title, trackCfg.VipMessage, trackCfg.Price, trackCfg.Duration)
		if err != nil {
			errs = errors.Join(errs, err)
		} else {
			tracks = append(tracks, track)
		}
	}

	if errs != nil {
		return nil, fmt.Errorf("failed to create tracks: %w", errs)
	}

	return tracks, nil
}
