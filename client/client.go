package client

import (
	"ATowerDefense/game"
	"fmt"
	"os"
	"time"

	"golang.org/x/term"
)

var viewOffsetX, viewOffsetY = 0, 0

func handleInput() error {
	in := make([]byte, 3)
	if _, err := os.Stdin.Read(in); err != nil {
		return err
	}

	if game.KeyBindContains(game.KeyBinds.Exit, in) {
		return game.Errors.Exit
	} else if game.KeyBindContains(game.KeyBinds.Up, in) {
		return nil
	} else if game.KeyBindContains(game.KeyBinds.Down, in) {
		return nil
	} else if game.KeyBindContains(game.KeyBinds.Right, in) {
		return nil
	} else if game.KeyBindContains(game.KeyBinds.Left, in) {
		return nil
	} else if game.KeyBindContains(game.KeyBinds.PanUp, in) {
		viewOffsetY -= 1
	} else if game.KeyBindContains(game.KeyBinds.PanDown, in) {
		viewOffsetY += 1
	} else if game.KeyBindContains(game.KeyBinds.PanRight, in) {
		viewOffsetX += 1
	} else if game.KeyBindContains(game.KeyBinds.PanLeft, in) {
		viewOffsetX -= 1
	} else if i := game.KeyBindIndex(game.KeyBinds.Numbers, in); i >= 0 {
		return nil
	}
	return nil
}

func visualize(gm *game.Game) {
	fmt.Print("\033[2J\033[0;0H")
	for y := range gm.GC.FieldHeight - viewOffsetY {
		if y+viewOffsetY < 0 {
			fmt.Print("\r\n")
			continue
		}
		for x := range gm.GC.FieldWidth - viewOffsetX {
			if x+viewOffsetX < 0 {
				fmt.Print(game.Green + "  " + game.Reset)
			} else if obj := gm.GetCollisions(x+viewOffsetX, y+viewOffsetY); len(obj) > 0 {
				switch obj[0].Type() {
				case "Field":
					fmt.Print(obj[0].Color() + "██" + game.Reset)
				case "Road":
					fmt.Print(obj[0].Color() + "██" + game.Reset)
				case "Tower":
					fmt.Print(obj[0].Color() + "██" + game.Reset)
				default:
					fmt.Print(obj[0].Color() + "??" + game.Reset)
				}
			} else {
				fmt.Print(game.Green + "██" + game.Reset)
			}
		}
		fmt.Print("\r\n")
	}
}

func Run(gc game.GameConfig) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return game.Errors.NotATerm
	}
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	gm := game.NewGame(gc)
	if err := gm.Start(); err != nil {
		return err
	}

	go func() {
		defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState); gm.Stop() }()

		for {
			err := handleInput()
			if err != nil {
				if err == game.Errors.Exit {
					break
				}
			}
		}
	}()

	for gm.GS.State != "stopped" {
		now := time.Now()

		gm.Iterate()
		visualize(gm)

		time.Sleep((time.Millisecond * 100) - time.Since(now))
	}

	return nil
}
