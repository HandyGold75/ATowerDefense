package client

import (
	"ATowerDefense/game"
	"errors"
	"fmt"
	"os"
	"slices"
	"time"

	"golang.org/x/term"
)

type (
	clientErrors struct{ NotATerm, Exit error }

	charSet string

	keybinds struct {
		Exit, Pause, Confirm, Delete,
		Up, Down, Right, Left,
		PanUp, PanDown, PanRight, PanLeft,
		SquereBracketLeft, SquereBracketRight,
		Plus, Minus,
		Numbers []keybind
	}
	keybind []byte
)

var (
	selectedX, selectedY     = 0, 0
	viewOffsetX, viewOffsetY = 0, 0
	selectedTower            = 0

	pid = 0

	tickDelay  = time.Millisecond * 50
	lagTracker = time.Duration(0)

	maxWidth, maxHeight = 0, 0

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

		// [
		SquereBracketLeft: []keybind{{91, 0, 0}},
		// ]
		SquereBracketRight: []keybind{{93, 0, 0}},

		// +
		Plus: []keybind{{43, 0, 0}},
		// -
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

func handleInput(gm *game.Game, debug bool) error {
	in := make([]byte, 3)
	if _, err := os.Stdin.Read(in); err != nil {
		return err
	}

	if keyBindContains(KeyBinds.Exit, in) {
		return Errors.Exit
	} else if debug {
		fmt.Printf("\033[2J\033[0;0H%v", in)
		return nil
	} else if keyBindContains(KeyBinds.Pause, in) {
		gm.TogglePause()
		return nil
	} else if keyBindContains(KeyBinds.Confirm, in) {
		if len(game.Towers) < selectedTower {
			return nil
		}
		err := gm.PlaceTower(game.Towers[selectedTower].Name, selectedX, selectedY, pid)
		if err != nil {
			if err == game.Errors.InvalidPlacement {
				return gm.DestoryTower(selectedX, selectedY, pid)
			}
			return err
		}

	} else if keyBindContains(KeyBinds.Delete, in) {
		return gm.StartRound()
	} else if keyBindContains(KeyBinds.Up, in) {
		selectedY = max(selectedY-1, max(0, viewOffsetY))
		return nil
	} else if keyBindContains(KeyBinds.Down, in) {
		selectedY = min(selectedY+1, min(gm.GC.FieldHeight, min(maxHeight, gm.GC.FieldHeight)+viewOffsetY)-1)
		return nil
	} else if keyBindContains(KeyBinds.Right, in) {
		selectedX = min(selectedX+1, min(gm.GC.FieldWidth, min(maxWidth, gm.GC.FieldWidth)+viewOffsetX)-1)
		return nil
	} else if keyBindContains(KeyBinds.Left, in) {
		selectedX = max(selectedX-1, max(0, viewOffsetX))
		return nil

	} else if keyBindContains(KeyBinds.PanUp, in) {
		viewOffsetY = max(viewOffsetY-1, -5)
		selectedY = max(selectedY-1, max(0, viewOffsetY))
		return nil
	} else if keyBindContains(KeyBinds.PanDown, in) {
		viewOffsetY = min(viewOffsetY+1, (gm.GC.FieldHeight-min(maxHeight, gm.GC.FieldHeight))+6)
		selectedY = min(selectedY+1, (gm.GC.FieldHeight+min(0, viewOffsetY))-1)
		return nil
	} else if keyBindContains(KeyBinds.PanRight, in) {
		viewOffsetX = min(viewOffsetX+1, (gm.GC.FieldWidth-min(maxWidth, gm.GC.FieldWidth))+5)
		selectedX = min(selectedX+1, (gm.GC.FieldWidth+min(0, viewOffsetX))-1)
		return nil
	} else if keyBindContains(KeyBinds.PanLeft, in) {
		viewOffsetX = max(viewOffsetX-1, -5)
		selectedX = max(selectedX-1, max(0, viewOffsetX))
		return nil

	} else if keyBindContains(KeyBinds.SquereBracketLeft, in) {
		selectedTower = max(selectedTower-1, 0)
		return nil
	} else if keyBindContains(KeyBinds.SquereBracketRight, in) {
		selectedTower = min(selectedTower+1, len(game.Towers)-1)
		return nil

	} else if keyBindContains(KeyBinds.Plus, in) {
		return nil
	} else if keyBindContains(KeyBinds.Minus, in) {
		return nil
	} else if i := keyBindIndex(KeyBinds.Numbers, in); i >= 0 {
		selectedTower = max(min(i, len(game.Towers)-1), 0)
		return nil

	}
	return nil
}

func Run(gc game.GameConfig, debug bool) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return Errors.NotATerm
	}
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	mw, mh, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return nil
	}
	maxWidth, maxHeight = int(mw/2), mh-1

	gm := game.NewGame(gc)
	if err := gm.Start(); err != nil {
		return err
	}
	pid = gm.AddPlayer()

	go func() {
		defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState); _ = gm.Stop() }()

		for {
			err := handleInput(gm, debug)
			if err != nil {
				if err == Errors.Exit {
					gm.GS.State = "stopped"
					break
				}
			}
		}
	}()

	last := time.Now()
	for gm.GS.State != "stopped" {
		now := time.Now()

		if !debug {
			gm.Iterate(time.Since(last))

			// if err := drawTui(gm); err != nil {
			// 	gm.GS.State = "stopped"
			// 	fmt.Println(err)
			// 	break
			// }

			if err := drawOpenGL(gm); err != nil {
				gm.GS.State = "stopped"
				fmt.Println(err)
				break
			}
		}

		last = now
		lagTracker = time.Since(now)
		time.Sleep(tickDelay - time.Since(now))
	}

	fmt.Print("\r\n")

	return nil
}
