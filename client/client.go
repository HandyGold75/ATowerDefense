package client

import (
	clsdl "ATowerDefense/client/sdl"
	cltui "ATowerDefense/client/tui"
	"ATowerDefense/game"
	"fmt"
	"time"
)

type (
	charSet string

	client interface {
		// Called in a defer, may be called multiple times.
		Stop()
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
	}
)

var (
	tickDelay   = time.Millisecond * 50
	processTime = time.Duration(0)
)

func Run(gc game.GameConfig, renderer string) error {
	gm := game.NewGame(gc)
	if err := gm.Start(); err != nil {
		return err
	}
	pid := gm.AddPlayer()

	var cl client = nil
	switch renderer {
	case "sdl":
		c, err := clsdl.NewSDL(gm, pid)
		if err != nil {
			return err
		}
		cl = c
	default:
		c, err := cltui.NewTUI(gm, pid)
		if err != nil {
			return err
		}
		cl = c
	}

	defer func() { cl.Stop(); _ = gm.Stop() }()
	go func() {
		defer func() { cl.Stop(); _ = gm.Stop() }()

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
