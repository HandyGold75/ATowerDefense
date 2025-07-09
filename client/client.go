package client

import (
	clsdl "ATowerDefense/client/sdl"
	cltui "ATowerDefense/client/tui"
	"ATowerDefense/game"
	"embed"
	"fmt"
	"time"
)

type (
	client interface {
		// Called after every game iterations.
		//
		// Time spent processing the previous game iteration and draw call is parsed to this function.
		// This does not include time spend waiting for the next tick cycle.
		//
		// Returning the `game.Errors.Exit` error will cause a succesfull game stop.
		// Any other error will cause a non succesfull game stop.
		Draw(time.Duration) error
		// Called in a goroutine until an error is returned.
		//
		// Returning the `game.Errors.Exit` error will cause a succesfull game stop.
		// Any other error will only be printed.
		Input() error
		// Called in a defer, may be called multiple times.
		Stop()
	}
)

var (
	processTime = time.Duration(0)

	//go:embed assets/*/*.png
	assets embed.FS
)

func Run(tui bool) error {
	gc := game.GameConfig{
		Mode:        "singleplayer",
		IP:          "84.25.253.77",
		Port:        17540,
		FieldHeight: 20,
		FieldWidth:  35,
		GameSpeed:   1,
		TickDelay:   time.Millisecond * 50,
	}

	var cl client = nil
	var gm *game.Game = nil
	if !tui {
		c, err := clsdl.NewSDL(gc, assets)
		if err != nil {
			return err
		}
		cl, gm = c, c.GM
	} else {
		c, err := cltui.NewTUI(gc)
		if err != nil {
			return err
		}
		cl, gm = c, c.GM
	}

	defer cl.Stop()
	go func() {
		defer cl.Stop()
		for gm.GS.State != "stopped" {
			if err := cl.Input(); err != nil {
				if err == game.Errors.Exit {
					break
				}
				fmt.Println(err)
			}
		}
	}()

	last := time.Now()
	for gm.GS.State != "stopped" {
		now := time.Now()

		gm.Iterate(time.Since(last))
		if err := cl.Draw(processTime); err != nil {
			if err == game.Errors.Exit {
				break
			}
			fmt.Println(err)
			break
		}

		last = now
		processTime = time.Since(now)
		time.Sleep(gm.GC.TickDelay - time.Since(now))
	}

	fmt.Print("\r\n")
	return nil
}
