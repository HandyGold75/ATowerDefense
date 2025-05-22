package client

import (
	"ATowerDefense/game"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

var viewOffsetX, viewOffsetY = 0, 0

func handleInput(gm *game.Game) error {
	in := make([]byte, 3)
	if _, err := os.Stdin.Read(in); err != nil {
		return err
	}

	if game.KeyBindContains(game.KeyBinds.Exit, in) {
		return game.Errors.Exit
	} else if game.KeyBindContains(game.KeyBinds.Pause, in) {
		_ = gm.TogglePause()
	} else if game.KeyBindContains(game.KeyBinds.Confirm, in) {
		_ = gm.StartRound()
	} else if game.KeyBindContains(game.KeyBinds.Delete, in) {
		return nil
	} else if game.KeyBindContains(game.KeyBinds.Up, in) {
		return nil
	} else if game.KeyBindContains(game.KeyBinds.Down, in) {
		return nil
	} else if game.KeyBindContains(game.KeyBinds.Right, in) {
		return nil
	} else if game.KeyBindContains(game.KeyBinds.Left, in) {
		return nil
	} else if game.KeyBindContains(game.KeyBinds.PanUp, in) {
		viewOffsetY = max(viewOffsetY-1, -5)
	} else if game.KeyBindContains(game.KeyBinds.PanDown, in) {
		_, maxHeight, _ := term.GetSize(int(os.Stdin.Fd()))
		viewOffsetY = min(viewOffsetY+1, (gm.GC.FieldHeight-min(maxHeight, gm.GC.FieldHeight))+5)
	} else if game.KeyBindContains(game.KeyBinds.PanRight, in) {
		maxWidth, _, _ := term.GetSize(int(os.Stdin.Fd()))
		viewOffsetX = min(viewOffsetX+1, (gm.GC.FieldWidth-min(int(maxWidth/2), gm.GC.FieldWidth))+5)
	} else if game.KeyBindContains(game.KeyBinds.PanLeft, in) {
		viewOffsetX = max(viewOffsetX-1, -5)
	} else if game.KeyBindContains(game.KeyBinds.Plus, in) {
		return nil
	} else if game.KeyBindContains(game.KeyBinds.Minus, in) {
		return nil
	} else if i := game.KeyBindIndex(game.KeyBinds.Numbers, in); i >= 0 {
		return nil
	}
	return nil
}

func drawBackground(maxWidth, maxHeight int, gm *game.Game) {
	for y := range min(gm.GC.FieldHeight, maxHeight) {
		if y != 0 {
			fmt.Print("\r\n")
		}
		if y+viewOffsetY < 0 || y+viewOffsetY >= gm.GC.FieldHeight {
			fmt.Print(strings.Repeat(string(game.BGBrightBlack+"  "+game.Reset), min(gm.GC.FieldWidth, int(maxWidth/2))))
			continue
		}
		for x := range min(gm.GC.FieldWidth, int(maxWidth/2)) {
			if x+viewOffsetX < 0 || x+viewOffsetX >= gm.GC.FieldWidth {
				fmt.Print(game.BGBrightBlack + "  " + game.Reset)
			} else if obj := gm.GetCollisions(x+viewOffsetX, y+viewOffsetY); len(obj) > 0 {
				switch obj[0].Type() {
				case "Field":
					fmt.Print(obj[0].Color() + "" + game.Reset)

				case "Road":
					switch obj[0].(*game.RoadObj).Direction {
					case "up":
						fmt.Print(obj[0].Color() + " " + game.Reset)
					case "right":
						fmt.Print(obj[0].Color() + " " + game.Reset)
					case "down":
						fmt.Print(obj[0].Color() + " " + game.Reset)
					case "left":
						fmt.Print(obj[0].Color() + " " + game.Reset)
					default:
						fmt.Print(obj[0].Color() + "?" + game.Reset)
					}

				case "Tower":
					fmt.Print(obj[0].Color() + " 󰚁" + game.Reset)

				default:
					fmt.Print(obj[0].Color() + "??" + game.Reset)
				}
			} else {
				fmt.Print(game.Green + "██" + game.Reset)
			}
		}
	}
}

func drawEnemies(maxWidth, maxHeight int, gm *game.Game) {
	for _, enemy := range gm.GS.Enemies {
		x, y := enemy.Cord()
		if x-viewOffsetX < 0 || x-viewOffsetX >= min(gm.GC.FieldWidth, int(maxWidth/2)) {
			return
		} else if y-viewOffsetY < 0 || y-viewOffsetY >= min(gm.GC.FieldHeight, maxHeight) {
			return
		}

		fmt.Print("\033[" + strconv.Itoa((y-viewOffsetY)+1) + ";" + strconv.Itoa(((x-viewOffsetX)*2)+1) + "H")
		fmt.Print(enemy.Color() + " " + game.Reset)
	}
}

func visualize(gm *game.Game) {
	fmt.Print("\033[2J\033[0;0H")
	maxWidth, maxHeight, _ := term.GetSize(int(os.Stdin.Fd()))

	drawBackground(maxWidth, maxHeight, gm)
	drawEnemies(maxWidth, maxHeight, gm)

	fmt.Print("\033[" + strconv.Itoa(maxHeight) + ";" + strconv.Itoa(maxWidth*2) + "H")
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
		defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState); _ = gm.Stop() }()

		for {
			err := handleInput(gm)
			if err != nil {
				if err == game.Errors.Exit {
					break
				}
			}
		}
	}()

	last := time.Now()
	for gm.GS.State != "stopped" {
		now := time.Now()

		gm.Iterate(time.Since(last))
		visualize(gm)

		last = now
		time.Sleep((time.Millisecond * 25) - time.Since(now))
	}

	fmt.Print("\r\n")

	return nil
}
