package display

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"github.com/nazarkurii/jukebox/internal/jukebox"
)

type CLI struct {
	jb  *jukebox.JukeBox
	in  io.Reader
	out io.Writer
}

func (cli CLI) write(s string) {
	_, err := cli.out.Write([]byte(s))
	if err != nil {
		log.Fatal(err)
	}
}

func (cli CLI) writef(s string, args ...any) {
	cli.write(fmt.Sprintf(s, args...))
}

func (cli *CLI) processUnexpectedError(err error) {
	log.Fatalf("An unexpected error occured %q, restart the jukebox manualy...", err)
}

func NewCLI(jb *jukebox.JukeBox, in io.Reader, out io.Writer) *CLI {
	return &CLI{jb: jb, in: in, out: out}
}

func (cli *CLI) Start() {
	cli.printTrackList()
	for {
		scanner := bufio.NewScanner(cli.in)
		if !scanner.Scan() {
			cli.processUnexpectedError(scanner.Err())
			continue
		}

		commandStr := strings.TrimSpace(scanner.Text())
		if commandStr == "" || strings.HasPrefix(commandStr, " ") {
			continue
		}

		switch cli.jb.State() {
		case jukebox.StateIdle:
			cli.processIdleCommand(commandStr)
		case jukebox.StateAcceptCoins:
			cli.processAcceptCoinsCommand(commandStr)
		default:
			cli.processUnexpectedError(errors.New("invalid jukebox state"))
		}
	}
}

func (cli *CLI) printTrackList() {
	tracks := cli.jb.TrackList()

	var builder strings.Builder

	builder.WriteString((`
╔══════════════════════════════════════════╗
║               🎵 JUKEBOX                 ║
╚══════════════════════════════════════════╝
	`))

	for _, t := range tracks {
		builder.WriteString(fmt.Sprintf("\n%d. %-45s %4s",
			t.Number,
			t.Title,
			formatMoney(t.Price),
		))
	}

	builder.WriteString("\nType track number or name to select\nType 'history' to view played tracks\n\nProvide track number: ")
	cli.write(builder.String())
}

func (cli *CLI) processIdleCommand(command string) {
	if command == "history" {
		cli.processHistoryCommand()
	} else {
		cli.processTrackChoosingCommand(command)
	}
}

func (cli *CLI) processHistoryCommand() {
	if history := cli.jb.History(); len(history) == 0 {
		cli.write("No tracks played yet\nProvide track number: ")
	} else {
		var builder strings.Builder
		for i, h := range history {
			builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, h))
		}
		builder.WriteString("\nProvide track number: ")
		cli.write(builder.String())
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
			cli.write("\nTrack not found\nProvide track number: ")
		} else {
			cli.processUnexpectedError(err)
		}
		return
	}

	cli.writef(`
Selected: %s %6s
Valid coin denominations: %v
Inserted: %s / Required: %s
Insert a coin: `,
		title, formatMoney(price), formatDenominations(cli.jb.CoinDenominations()), formatMoney(0), formatMoney(price))
}

func (cli *CLI) processTrackCancelation() {
	change := cli.jb.CancelTrack()
	var builder strings.Builder

	builder.WriteString("Track canceled")

	if change != nil {
		builder.WriteString("\nChange:" + formatCoins(change))
	}

	cli.write(builder.String())
}

func (cli *CLI) processCoinAcceptance(command string) {
	providedCoin, err := strconv.ParseFloat(command, 64)
	if err != nil {
		cli.writef("\nInvalid coin format (example: %.2f)\nInsert a coin: ", float64(cli.jb.CoinDenominations()[1]/100))
		return
	}

	totalAccepted, price, err := cli.jb.AcceptCoin(int(providedCoin * 100))
	if err != nil {
		if errors.Is(err, jukebox.ErrInvalidCoinDenomination) {
			cli.writef(`
Invalid coin denomination
Allowed: %v
Insert a coin: `, formatDenominations(cli.jb.CoinDenominations()))

		} else {
			cli.processUnexpectedError(err)
		}
		return
	}

	if totalAccepted >= price {
		title, err := cli.jb.GetChosenTrackTitle()
		if err != nil {
			cli.processUnexpectedError(err)
			return
		}

		cli.write("\nEnough, starting to play the track...\nPlaying: " + title)

		change, err := cli.jb.PlayChosenTrack()
		if err != nil {
			if errors.Is(err, jukebox.ErrImpossibleChange) {
				cli.write("We can't calculate yout change due to lack of funds, please contact support for refund...")
			}
			cli.write("Something went wrong... \nHere is your refund: " + formatCoins(change))
			cli.processUnexpectedError(err)
			return
		}

		msg := "\nDone!"
		if change != nil {
			msg += " Change: " + formatCoins(change)
		}
		cli.write(msg)

		cli.printTrackList()
	} else {
		cli.writef("Inserted: %s / Required: %s\nInsert a coin: ", formatMoney(totalAccepted), formatMoney(price))
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
