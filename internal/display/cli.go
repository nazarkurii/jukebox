package display

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/nazarkurii/jukebox/internal/jukebox"
)

type CLI struct {
	jb *jukebox.JukeBox
}

func NewCLI(jb *jukebox.JukeBox) *CLI {
	return &CLI{jb: jb}
}

func (cli *CLI) Start() {
	cli.printTrackList()

	for {
		var command string
		fmt.Scan(&command)

		switch cli.jb.State() {
		case jukebox.StateIdle:
			cli.processIdleCommand(command)
		case jukebox.StateAcceptCoins:
			cli.processAcceptCoinsCommand(command)
		default:
			fmt.Println("Uexpected error, restarting the jukebox...")
			cli.printTrackList()
		}
	}
}

func (cli *CLI) printTrackList() {
	tracks := cli.jb.TrackList()

	fmt.Println(`
╔══════════════════════════════════════════╗
║               🎵 JUKEBOX                 ║
╚══════════════════════════════════════════╝
	`)

	for _, t := range tracks {
		fmt.Printf("%d. %-45s %4s\n",
			t.Number,
			t.Title,
			formatMoney(t.Price),
		)
	}

	fmt.Printf("\nType track number or name to select\nType 'history' to view played tracks\n\nProvide track number:")
}

func (cli *CLI) processUnexpectedError(_ error) {
	fmt.Printf("An unexpected error occured, restarting the jukebox...\n\n")
	cli.printTrackList()
}

func (cli *CLI) processIdleCommand(command string) {
	if command == "history" {
		cli.processHistoryCommand()
		fmt.Printf("Provide track number:")
	} else {
		cli.processTrackChoosingCommand(command)
	}
}

func (cli *CLI) processHistoryCommand() {
	defer fmt.Println()

	if history := cli.jb.History(); len(history) == 0 {
		fmt.Println("No tracks played yet")
	} else {
		for i, h := range history {
			fmt.Printf("%d. %s\n", i+1, h)
		}
	}
}

func (cli *CLI) processTrackChoosingCommand(command string) {
	trackNumber, err := strconv.Atoi(command)

	var title string
	var price int

	if err != nil {
		title, price, err = cli.jb.ChooseTrackByName(command)
	} else {
		title, price, err = cli.jb.ChooseTrackByNumber(trackNumber)
	}

	if err != nil {
		if errors.Is(err, jukebox.ErrTrackNotFound) {
			fmt.Printf("\nTrack not found\nProvide track number:")
		} else {
			cli.processUnexpectedError(err)
		}
		return
	}

	fmt.Printf("\nSelected: %s %6s\n", title, formatMoney(price))
	fmt.Println("\nValid coin denominations: ", formatDenominations(cli.jb.CoinDenominations()))
	fmt.Printf("Inserted: %s / Required: %s\nInsert a coin: ", formatMoney(0), formatMoney(price))
}

func (cli *CLI) processTrackCancelation() {
	change := cli.jb.CancelTrack()
	fmt.Println("Track canceled")

	if change != nil {
		fmt.Println("Change:", formatCoins(change))
	}
}

func (cli *CLI) processCoinAcceptance(command string) {
	providedCoin, err := strconv.ParseFloat(command, 64)
	if err != nil {
		fmt.Printf("\nInvalid coin format (example: %.2f)\nInsert a coin: ", float64(cli.jb.CoinDenominations()[1]/100))
		return
	}

	totalAccepted, price, err := cli.jb.AcceptCoin(int(providedCoin * 100))
	if err != nil {
		if errors.Is(err, jukebox.ErrInvalidCoinDenomination) {
			fmt.Println("Invalid coin denomination")
			fmt.Println("Allowed:", formatDenominations(cli.jb.CoinDenominations()))
			fmt.Printf("Insert a coin: ")
		} else {
			cli.processUnexpectedError(err)
		}
		return
	}

	if totalAccepted >= price {
		title, play, err := cli.jb.PlayChosenTrack()
		if err != nil {
			cli.processUnexpectedError(err)
			return
		}

		fmt.Println("\nEnough, starting to play the track...\nPlaying: ", title)

		change, err := play()
		if err != nil {
			if errors.Is(err, jukebox.ErrImpossibleChange) {
				fmt.Println("We can't calculate yout change due to lack of funds, please contact support for refund...")
			}
			fmt.Println("Something went wrong... \nHere is your refund: " + formatCoins(change))
			return
		}

		msg := "\nDone!"
		if change != nil {
			msg += " Change: " + formatCoins(change)
		}
		fmt.Println(msg)

		cli.printTrackList()
	} else {
		fmt.Printf("Inserted: %s / Required: %s\nInsert a coin: ", formatMoney(totalAccepted), formatMoney(price))
	}
}

func (cli *CLI) processAcceptCoinsCommand(command string) {
	if command == "cancel" {
		cli.processTrackCancelation()
		cli.printTrackList()
	} else {
		cli.processCoinAcceptance(command)
	}
}

func formatCoins(coins []int) string {
	var parts []string
	for _, c := range coins {
		parts = append(parts, formatMoney(c))
	}
	return strings.Join(parts, ", ")
}

func formatDenominations(coins []int) string {
	var parts []string
	for _, c := range coins {
		parts = append(parts, formatMoney(c))
	}
	return strings.Join(parts, ", ")
}

func formatMoney(cents int) string {
	return fmt.Sprintf("$%.2f", float64(cents)/100)
}
