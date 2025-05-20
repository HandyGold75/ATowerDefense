package main

import (
	"ATowerDefense/client"
	"ATowerDefense/game"
	"fmt"
	"os"
	"strconv"

	"github.com/HandyGold75/GOLib/argp"
	"github.com/HandyGold75/GOLib/tui"
)

var args = argp.ParseArgs(struct {
	Help   bool   `switch:"h,-help" opts:"help"        help:"Another game of Snake."`
	Server bool   `switch:"s,-server"                  help:"Start as a server instace."`
	IP     string `switch:"i,-ip" default:"0.0.0.0"    help:"Listen on this ip when started as server."`
	Port   uint16 `switch:"p,-port" default:"17540"    help:"Listen on this port when started as server."`
}{})

func menu() (gc game.GameConfig, err error) {
	mode := ""

	tui.Defaults.Align = tui.AlignLeft
	mm := tui.NewMenuBulky("ASnake")

	sp := mm.Menu.NewMenu("SinglePlayer")
	sp.NewAction("Start", func() { mode = "singleplayer" })
	spFieldHeight := sp.NewDigit("Field height", 50, 10, 9999)
	spFieldWidth := sp.NewDigit("Field width", 50, 10, 9999)

	mp := mm.Menu.NewMenu("MultiPlayer")
	mp.NewAction("Connect", func() { mode = "multiplayer" })
	mpIP := mp.NewIPv4("IP", "84.25.253.77")
	mpPort := mp.NewDigit("Port", 17540, 0, 65535)

	if err := mm.Run(); err != nil {
		return gc, err
	}

	gc.Mode = mode
	gc.IP = mpIP.Value()

	port, err := strconv.ParseUint(mpPort.Value(), 10, 16)
	if err != nil {
		return gc, err
	}
	gc.Port = uint16(port)

	if gc.FieldHeight, err = strconv.Atoi(spFieldHeight.Value()); err != nil {
		return gc, err
	}
	if gc.FieldWidth, err = strconv.Atoi(spFieldWidth.Value()); err != nil {
		return gc, err
	}
	return gc, nil
}

func main() {
	// if args.Server {
	// 	if err := server.Run(args.IP, args.Port); err != nil {
	// 		fmt.Println(err)
	// 		os.Exit(1)
	// 	}
	// 	os.Exit(0)
	// }

	// gc, err := menu()
	// if err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	// if err := client.Run(gc); err != nil {
	// 	fmt.Println(err)
	// 	os.Exit(1)
	// }

	if err := client.Run(game.GameConfig{
		Mode:        "singleplayer",
		IP:          "84.25.253.77",
		Port:        17540,
		FieldHeight: 25,
		FieldWidth:  25,
	}); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
