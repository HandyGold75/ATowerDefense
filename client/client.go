package client

import (
	"ATowerDefense/game"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"golang.org/x/term"
)

type (
	clientErrors struct{ NotATerm, Exit error }

	charSet string

	keybinds struct{ Up, Down, Right, Left, PanUp, PanDown, PanRight, PanLeft, Plus, Minus, Exit, Pause, Confirm, Delete, Numbers []keybind }
	keybind  []byte
)

var (
	viewOffsetX, viewOffsetY = 0, 0
	selectedTower            = 0

	KeyBinds = keybinds{
		// ESC, CTRL_C, CTRL_D,
		Exit: []keybind{{27, 0, 0}, {3, 0, 0}, {4, 0, 0}},
		// P, Q
		Pause: []keybind{{112, 0, 0}, {113, 0, 0}},
		// RETURN
		Confirm: []keybind{{13, 0, 0}},
		// BACKSPACE, DEL
		Delete: []keybind{{127, 0, 0}, {27, 91, 51}},

		// W, K
		Up: []keybind{{119, 0, 0}, {107, 0, 0}},
		// S, J
		Down: []keybind{{115, 0, 0}, {106, 0, 0}},
		// D, L
		Right: []keybind{{100, 0, 0}, {108, 0, 0}},
		// A, H
		Left: []keybind{{97, 0, 0}, {104, 0, 0}},

		// UP
		PanUp: []keybind{{27, 91, 65}},
		// DOWN
		PanDown: []keybind{{27, 91, 66}},
		// RIGHT,
		PanRight: []keybind{{27, 91, 67}},
		// LEFT,
		PanLeft: []keybind{{27, 91, 68}},

		// PLUS
		Plus: []keybind{{43, 0, 0}},
		// MINUS
		Minus: []keybind{{45, 0, 0}},

		// 0, 1, 2, 3, 4, 5, 6, 7, 8, 9
		Numbers: []keybind{{48, 0, 0}, {49, 0, 0}, {50, 0, 0}, {51, 0, 0}, {52, 0, 0}, {53, 0, 0}, {54, 0, 0}, {55, 0, 0}, {56, 0, 0}, {57, 0, 0}},
	}

	Errors = clientErrors{
		NotATerm: errors.New("stdin/ stdout should be a terminal"),
		Exit:     errors.New("game is exiting"),
	}
)

const (
	Letters        charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Digits         charSet = "0123456789"
	Hex            charSet = "0123456789abcdefABCDEF"
	WhiteSpace     charSet = " "
	Punctuation    charSet = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
	GeneralCharSet charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
)

func keyBindContains(kb []keybind, b []byte) bool {
	return slices.ContainsFunc(kb, func(v keybind) bool { return slices.Equal(v, b) })
}

func keyBindIndex(kb []keybind, b []byte) int {
	return slices.IndexFunc(kb, func(v keybind) bool { return slices.Equal(v, b) })
}

func handleInput(gm *game.Game) error {
	in := make([]byte, 3)
	if _, err := os.Stdin.Read(in); err != nil {
		return err
	}

	if keyBindContains(KeyBinds.Exit, in) {
		return Errors.Exit
	} else if keyBindContains(KeyBinds.Pause, in) {
		_ = gm.TogglePause()
	} else if keyBindContains(KeyBinds.Confirm, in) {
		_ = gm.StartRound()
	} else if keyBindContains(KeyBinds.Delete, in) {
		return nil
	} else if keyBindContains(KeyBinds.Up, in) {
		return nil
	} else if keyBindContains(KeyBinds.Down, in) {
		return nil
	} else if keyBindContains(KeyBinds.Right, in) {
		return nil
	} else if keyBindContains(KeyBinds.Left, in) {
		return nil
	} else if keyBindContains(KeyBinds.PanUp, in) {
		viewOffsetY = max(viewOffsetY-1, -5)
	} else if keyBindContains(KeyBinds.PanDown, in) {
		_, maxHeight, _ := term.GetSize(int(os.Stdin.Fd()))
		viewOffsetY = min(viewOffsetY+1, (gm.GC.FieldHeight-min(maxHeight, gm.GC.FieldHeight))+5)
	} else if keyBindContains(KeyBinds.PanRight, in) {
		maxWidth, _, _ := term.GetSize(int(os.Stdin.Fd()))
		viewOffsetX = min(viewOffsetX+1, (gm.GC.FieldWidth-min(int(maxWidth/2), gm.GC.FieldWidth))+5)
	} else if keyBindContains(KeyBinds.PanLeft, in) {
		viewOffsetX = max(viewOffsetX-1, -5)
	} else if keyBindContains(KeyBinds.Plus, in) {
		selectedTower = min(selectedTower+1, len(game.Towers)-1)
	} else if keyBindContains(KeyBinds.Minus, in) {
		selectedTower = max(selectedTower-1, 0)
	} else if i := keyBindIndex(KeyBinds.Numbers, in); i >= 0 {
		return nil
	}
	return nil
}

func drawField(maxWidth, maxHeight int, gm *game.Game) {
	fmt.Print("\033[0;0H")
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
				switch obj[len(obj)-1].Type() {
				case "Obstacle":
					fmt.Print(obj[len(obj)-1].Color() + "" + game.Reset)

				case "Road":
					switch obj[len(obj)-1].(*game.RoadObj).Direction {
					case "up":
						fmt.Print(obj[len(obj)-1].Color() + " " + game.Reset)
					case "right":
						fmt.Print(obj[len(obj)-1].Color() + " " + game.Reset)
					case "down":
						fmt.Print(obj[len(obj)-1].Color() + " " + game.Reset)
					case "left":
						fmt.Print(obj[len(obj)-1].Color() + " " + game.Reset)
					default:
						fmt.Print(obj[len(obj)-1].Color() + "?" + game.Reset)
					}

				case "Tower":
					fmt.Print(obj[len(obj)-1].Color() + " 󰚁" + game.Reset)

				case "Enemy":
					fmt.Print(obj[len(obj)-1].Color() + " " + game.Reset)

				default:
					fmt.Print(obj[len(obj)-1].Color() + "??" + game.Reset)
				}
			} else {
				fmt.Print(game.Green + "██" + game.Reset)
			}
		}
	}
}

func drawUI(maxWidth, maxHeight int, gm *game.Game) {
	fmt.Print("\033[0;0H" + string(game.BGBrightBlack) +
		string(game.Red) + " [" + strconv.Itoa(gm.GS.Health) + "] " +
		string(game.White) + " " + gm.GS.Phase + " ")
	if gm.GS.Phase == "defending" {
		fmt.Print(string(game.White) + "(" + strconv.Itoa(len(gm.GS.Enemies)) + ") ")
	}
	fmt.Print(string(game.Reset))

	if maxWidth*2 > gm.GC.FieldWidth {
		for i, tower := range game.Towers {
			fmt.Print("\033[" + strconv.Itoa(i+1) + ";" + strconv.Itoa((gm.GC.FieldWidth*2)+1) + "H")
			if i == selectedTower {
				fmt.Print(string(game.BGWhite+game.Black) + tower.Name + string(game.Reset))
			} else {
				fmt.Print(string(game.BGBlack+game.White) + tower.Name + string(game.Reset))
			}
		}
	}
}

// func drawEnemies(maxWidth, maxHeight int, gm *game.Game) {
// 	for _, enemy := range gm.GS.Enemies {
// 		x, y := enemy.Cord()
// 		if x-viewOffsetX < 0 || x-viewOffsetX >= min(gm.GC.FieldWidth, int(maxWidth/2)) {
// 			return
// 		} else if y-viewOffsetY < 0 || y-viewOffsetY >= min(gm.GC.FieldHeight, maxHeight) {
// 			return
// 		}

// 		fmt.Print("\033[" + strconv.Itoa((y-viewOffsetY)+1) + ";" + strconv.Itoa(((x-viewOffsetX)*2)+1) + "H")
// 		fmt.Print(enemy.Color() + " " + game.Reset)
// 	}
// }

func visualize(gm *game.Game) {
	fmt.Print("\033[2J")
	maxWidth, maxHeight, _ := term.GetSize(int(os.Stdin.Fd()))

	drawField(maxWidth, maxHeight, gm)
	drawUI(maxWidth, maxHeight, gm)

	fmt.Print("\033[" + strconv.Itoa(maxHeight) + ";" + strconv.Itoa(maxWidth*2) + "H")
}

func Run(gc game.GameConfig) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return Errors.NotATerm
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
				if err == Errors.Exit {
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
