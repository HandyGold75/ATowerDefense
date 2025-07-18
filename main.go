package main

import (
	clsdl "ATowerDefense/client/sdl"
	cltui "ATowerDefense/client/tui"
	"ATowerDefense/game"
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/HandyGold75/GOLib/argp"
)

var (
	args = argp.ParseArgs(struct {
		Help             bool    `switch:"h,-help"              opts:"help"   help:"Another game of Snake."`
		FieldWidth       int     `switch:"w,-field-width"       default:"35"  help:"Game setting: Field Width"`
		FieldHeight      int     `switch:"h,-field-height"      default:"20"  help:"Game setting: Field Height"`
		RefundMultiplier float64 `switch:"r,-refund-multiplier" default:"0.8" help:"Game setting: Refund Multiplier"`
		TUI              bool    `switch:"t,-tui"                             help:"Use TUI renderer"`
	}{})

	//go:embed assets/*/*.png
	assets embed.FS
)

func main() {
	gc := game.GameConfig{
		FieldHeight:      args.FieldHeight,
		FieldWidth:       args.FieldWidth,
		GameSpeed:        1,
		RefundMultiplier: args.RefundMultiplier,
		TickDelay:        time.Millisecond * 50,
	}

	if args.TUI {
		if err := cltui.Run(gc); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else {
		if err := clsdl.Run(gc, assets); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}
}
