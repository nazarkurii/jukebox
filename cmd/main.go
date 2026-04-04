package main

import (
	"os"

	"github.com/nazarkurii/jukebox/config"
	coins "github.com/nazarkurii/jukebox/internal/coins"
	"github.com/nazarkurii/jukebox/internal/display"
	"github.com/nazarkurii/jukebox/internal/jukebox"
)

func main() {
	tracks, err := config.LoadTracksFromJSON(os.Getenv("JUKEBOX__TRACKS_PATH"))
	if err != nil {
		panic("failed to load config tracks: " + err.Error())
	}

	display.NewCLI(jukebox.New(tracks, coins.NewPolicy(1, 5, 10, 25, 50, 100))).Start()
}
