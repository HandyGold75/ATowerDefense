package client

import (
	clsdl "ATowerDefense/client/sdl"
	cltui "ATowerDefense/client/tui"
	"ATowerDefense/game"
	"embed"
	"fmt"
	"strconv"
	"time"

	"github.com/HandyGold75/GOLib/tui"
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

	//go:embed assets/*.png
	assets embed.FS
)

func menuTUI(gc game.GameConfig) (game.GameConfig, error) {
	mode := ""

	tui.Defaults.Align = tui.AlignLeft
	mm := tui.NewMenuBulky("ASnake")

	sp := mm.Menu.NewMenu("SinglePlayer")
	sp.NewAction("Start", func() { mode = "singleplayer" })
	spFieldWidth := sp.NewDigit("Field width", gc.FieldWidth, 10, 9999)
	spFieldHeight := sp.NewDigit("Field height", gc.FieldHeight, 10, 9999)

	mp := mm.Menu.NewMenu("MultiPlayer")
	mp.NewAction("Connect", func() { mode = "multiplayer" })
	mpIP := mp.NewIPv4("IP", gc.IP)
	mpPort := mp.NewDigit("Port", int(gc.Port), 0, 65535)

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

	if tui {
		conf, err := menuTUI(gc)
		if err != nil {
			return err
		}
		gc = conf
	}

	gm := game.NewGame(gc)
	if err := gm.Start(); err != nil {
		return err
	}
	pid := gm.AddPlayer()

	var cl client = nil
	if !tui {
		c, err := clsdl.NewSDL(gm, pid, assets)
		if err != nil {
			return err
		}
		cl = c
	} else {
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
