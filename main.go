package main

import (
	"ATowerDefense/client"
	"ATowerDefense/server"
	"fmt"
	"os"

	"github.com/HandyGold75/GOLib/argp"
)

var args = argp.ParseArgs(struct {
	Help   bool   `switch:"h,-help" opts:"help"        help:"Another game of Snake."`
	Server bool   `switch:"s,-server"                  help:"Start as a server instace."`
	IP     string `switch:"i,-ip" default:"0.0.0.0"    help:"Listen on this ip when started as server."`
	Port   uint16 `switch:"p,-port" default:"17540"    help:"Listen on this port when started as server."`
	TUI    bool   `switch:"t,-tui"                  help:"Use TUI renderer"`
}{})

func main() {
	if args.Server {
		if err := server.Run(args.IP, args.Port); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	if err := client.Run(args.TUI); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
