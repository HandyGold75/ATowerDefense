package main

import (
	clsdl "ATowerDefense/client/sdl"
	cltui "ATowerDefense/client/tui"
	"ATowerDefense/game"
	"ATowerDefense/server"
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/HandyGold75/GOLib/argp"
)

var (
	args = argp.ParseArgs(struct {
		Help   bool   `switch:"h,-help" opts:"help"        help:"Another game of Snake."`
		Server bool   `switch:"s,-server"                  help:"Start as a server instace."`
		IP     string `switch:"i,-ip" default:"0.0.0.0"    help:"Listen on this ip when started as server."`
		Port   uint16 `switch:"p,-port" default:"17540"    help:"Listen on this port when started as server."`
		TUI    bool   `switch:"t,-tui"                     help:"Use TUI renderer"`
	}{})

	//go:embed client/assets/*/*.png
	assets embed.FS
)

func main() {
	gc := game.GameConfig{
		Mode:        "singleplayer",
		IP:          "84.25.253.77",
		Port:        17540,
		FieldHeight: 20,
		FieldWidth:  35,
		GameSpeed:   1,
		TickDelay:   time.Millisecond * 50,
	}

	if args.Server {
		gc.IP = args.IP
		gc.Port = args.Port
		if err := server.Run(gc); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	} else if args.TUI {
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
