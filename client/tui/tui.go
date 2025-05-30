package cltui

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
	keybinds struct {
		Exit, Pause, Confirm, Delete,
		Up, Down, Right, Left,
		PanUp, PanDown, PanRight, PanLeft,
		SquereBracketLeft, SquereBracketRight,
		Plus, Minus,
		Numbers []keybind
	}
	keybind []byte

	TUI struct {
		game *game.Game
		pid  int

		oldState *term.State

		selectedX, selectedY,
		viewOffsetX, viewOffsetY,
		selectedTower int

		maxWidth, maxHeight int

		keyBinds keybinds
	}
)

func NewTUI(gm *game.Game, pid int) (*TUI, error) {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return nil, errors.New("stdin is not a terminal")
	}
	state, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}
	mw, mh, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return nil, err
	}

	return &TUI{
		game: gm, pid: pid,
		oldState: state,

		selectedX: 0, selectedY: 0,
		viewOffsetX: 0, viewOffsetY: 0,
		selectedTower: 0,

		maxWidth: int(mw / 2), maxHeight: mh - 1,

		keyBinds: keybinds{
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
		},
	}, nil
}

func (cl *TUI) Draw(processTime time.Duration) error {
	mw, mh, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	cl.maxWidth, cl.maxHeight = int(mw/2), mh-1

	fmt.Print("\033[2J" + cl.getField() + cl.getUI(processTime) + "\033[" + strconv.Itoa(cl.maxHeight) + ";" + strconv.Itoa(cl.maxWidth*2) + "H")
	return nil
}

func (cl *TUI) Stop() {
	if cl.oldState != nil {
		_ = term.Restore(int(os.Stdin.Fd()), cl.oldState)
		cl.oldState = nil
	}
}

func (cl *TUI) Input() error {
	in := make([]byte, 3)
	if _, err := os.Stdin.Read(in); err != nil {
		return err
	}

	if keyBindContains(cl.keyBinds.Exit, in) {
		return game.Errors.Exit
	} else if keyBindContains(cl.keyBinds.Pause, in) {
		cl.game.TogglePause()
		return nil
	} else if keyBindContains(cl.keyBinds.Confirm, in) {
		if len(game.Towers) < cl.selectedTower {
			return nil
		}
		err := cl.game.PlaceTower(game.Towers[cl.selectedTower].Name, cl.selectedX, cl.selectedY, cl.pid)
		if err != nil {
			if err == game.Errors.InvalidPlacement {
				return cl.game.DestoryTower(cl.selectedX, cl.selectedY, cl.pid)
			}
			return err
		}

	} else if keyBindContains(cl.keyBinds.Delete, in) {
		return cl.game.StartRound()
	} else if keyBindContains(cl.keyBinds.Up, in) {
		cl.selectedY = max(cl.selectedY-1, max(0, cl.viewOffsetY))
		return nil
	} else if keyBindContains(cl.keyBinds.Down, in) {
		cl.selectedY = min(cl.selectedY+1, min(cl.game.GC.FieldHeight, min(cl.maxHeight, cl.game.GC.FieldHeight)+cl.viewOffsetY)-1)
		return nil
	} else if keyBindContains(cl.keyBinds.Right, in) {
		cl.selectedX = min(cl.selectedX+1, min(cl.game.GC.FieldWidth, min(cl.maxWidth, cl.game.GC.FieldWidth)+cl.viewOffsetX)-1)
		return nil
	} else if keyBindContains(cl.keyBinds.Left, in) {
		cl.selectedX = max(cl.selectedX-1, max(0, cl.viewOffsetX))
		return nil

	} else if keyBindContains(cl.keyBinds.PanUp, in) {
		cl.viewOffsetY = max(cl.viewOffsetY-1, -5)
		cl.selectedY = max(cl.selectedY-1, max(0, cl.viewOffsetY))
		return nil
	} else if keyBindContains(cl.keyBinds.PanDown, in) {
		cl.viewOffsetY = min(cl.viewOffsetY+1, (cl.game.GC.FieldHeight-min(cl.maxHeight, cl.game.GC.FieldHeight))+6)
		cl.selectedY = min(cl.selectedY+1, (cl.game.GC.FieldHeight+min(0, cl.viewOffsetY))-1)
		return nil
	} else if keyBindContains(cl.keyBinds.PanRight, in) {
		cl.viewOffsetX = min(cl.viewOffsetX+1, (cl.game.GC.FieldWidth-min(cl.maxWidth, cl.game.GC.FieldWidth))+5)
		cl.selectedX = min(cl.selectedX+1, (cl.game.GC.FieldWidth+min(0, cl.viewOffsetX))-1)
		return nil
	} else if keyBindContains(cl.keyBinds.PanLeft, in) {
		cl.viewOffsetX = max(cl.viewOffsetX-1, -5)
		cl.selectedX = max(cl.selectedX-1, max(0, cl.viewOffsetX))
		return nil

	} else if keyBindContains(cl.keyBinds.SquereBracketLeft, in) {
		cl.selectedTower = max(cl.selectedTower-1, 0)
		return nil
	} else if keyBindContains(cl.keyBinds.SquereBracketRight, in) {
		cl.selectedTower = min(cl.selectedTower+1, len(game.Towers)-1)
		return nil

	} else if keyBindContains(cl.keyBinds.Plus, in) {
		return nil
	} else if keyBindContains(cl.keyBinds.Minus, in) {
		return nil
	} else if i := keyBindIndex(cl.keyBinds.Numbers, in); i >= 0 {
		cl.selectedTower = max(min(i, len(game.Towers)-1), 0)
		return nil

	}

	return nil
}

func keyBindContains(kb []keybind, b []byte) bool {
	return slices.ContainsFunc(kb, func(v keybind) bool { return slices.Equal(v, b) })
}

func keyBindIndex(kb []keybind, b []byte) int {
	return slices.IndexFunc(kb, func(v keybind) bool { return slices.Equal(v, b) })
}

func (cl *TUI) getField() string {
	frame := "\033[2;0H"
	for y := range min(cl.game.GC.FieldHeight, cl.maxHeight) {
		if y != 0 {
			frame += "\r\n"
		}
		if y+cl.viewOffsetY < 0 || y+cl.viewOffsetY >= cl.game.GC.FieldHeight {
			frame += strings.Repeat(string(game.BGBrightBlack+"  "+game.Reset), min(cl.game.GC.FieldWidth, cl.maxWidth))
			continue
		}
		for x := range min(cl.game.GC.FieldWidth, cl.maxWidth) {
			if x+cl.viewOffsetX < 0 || x+cl.viewOffsetX >= cl.game.GC.FieldWidth {
				frame += string(game.BGBrightBlack + "  " + game.Reset)
			} else if x+cl.viewOffsetX == cl.selectedX && y+cl.viewOffsetY == cl.selectedY {
				frame += string(game.BGGreen + game.Black + "" + game.Reset)
			} else if obj := cl.game.GetCollisions(x+cl.viewOffsetX, y+cl.viewOffsetY); len(obj) > 0 {
				switch obj[len(obj)-1].Type() {
				case "Obstacle":
					frame += string(obj[len(obj)-1].Color() + "" + game.Reset)

				case "Road":
					if obj[len(obj)-1].(*game.RoadObj).Index == 0 {
						frame += string(obj[len(obj)-1].Color() + game.BrightBlack + " 󰮢" + game.Reset)
						continue
					} else if obj[len(obj)-1].(*game.RoadObj).Index == len(cl.game.GS.Roads)-1 {
						frame += string(obj[len(obj)-1].Color() + game.BrightBlack + " 󰄚" + game.Reset)
						continue
					}

					switch obj[len(obj)-1].(*game.RoadObj).Direction {
					case "up":
						frame += string(obj[len(obj)-1].Color() + " " + game.Reset)
					case "right":
						frame += string(obj[len(obj)-1].Color() + " " + game.Reset)
					case "down":
						frame += string(obj[len(obj)-1].Color() + " " + game.Reset)
					case "left":
						frame += string(obj[len(obj)-1].Color() + " " + game.Reset)
					default:
						frame += string(obj[len(obj)-1].Color() + "?" + game.Reset)
					}

				case "Tower":
					frame += string(obj[len(obj)-1].Color() + " 󰚁" + game.Reset)

				case "Enemy":
					roadX, roadY := cl.game.GS.Roads[0].Cord()
					enemyX, enemyY := obj[len(obj)-1].Cord()
					if enemyX == roadX && enemyY == roadY {
						frame += string(obj[len(obj)-1].Color() + game.BrightBlack + " 󰮢" + game.Reset)
						continue
					}
					frame += string(obj[len(obj)-1].Color() + " " + game.Reset)

				default:
					frame += string(obj[len(obj)-1].Color() + "??" + game.Reset)
				}
			} else {
				frame += string(game.Green + "██" + game.Reset)
			}
		}
	}
	return frame
}

func (cl *TUI) getUI(processTime time.Duration) string {
	state := cl.game.GS.Phase
	if cl.game.GS.State == "paused" {
		state += " [p]"
	}
	phase := "R:" + strconv.Itoa(cl.game.GS.Round+1)
	if cl.game.GS.Phase == "defending" {
		phase = "E:" + strconv.Itoa(len(cl.game.GS.Enemies))
	}
	msgLen := len(state) + len(phase) + 1
	msgLeft := fmt.Sprintf(string(game.BrightWhite)+"%v %v", state, phase)

	lag := strconv.FormatInt(processTime.Milliseconds(), 10)
	if processTime >= cl.game.GC.TickDelay {
		msgLen -= 4
		lag = string(game.Red) + lag
	}
	msgLen += len(lag) + len(strconv.Itoa(cl.game.Players[cl.pid].Coins)) + len(strconv.Itoa(cl.game.GS.Health)) + 2
	msgRight := fmt.Sprintf(string(game.White)+"%v "+
		string(game.BrightYellow)+"%v "+
		string(game.BrightRed)+"%v", lag, cl.game.Players[cl.pid].Coins, cl.game.GS.Health)

	frame := fmt.Sprintf("\033[0;0H"+string(game.BGBrightBlack)+"%v"+strings.Repeat(" ", max(1, min(cl.game.GC.FieldWidth*2, cl.maxWidth*2)-msgLen))+"%v"+string(game.Reset), msgLeft, msgRight)

	if cl.maxWidth > cl.game.GC.FieldWidth+10 && cl.maxHeight+1 >= len(game.Towers) {
		for i, tower := range game.Towers {
			frame += "\033[" + strconv.Itoa(i+1) + ";" + strconv.Itoa((cl.game.GC.FieldWidth*2)+1) + "H"
			msgLeft := tower.Name
			msgRight := "(" + strconv.Itoa(tower.Cost) + ")"
			if i == cl.selectedTower {
				frame += string(game.BGWhite+game.Black) + msgLeft + strings.Repeat(" ", max(0, 20-len(msgLeft)-len(msgRight))) + msgRight + string(game.Reset)
			} else {
				frame += string(game.BGBlack+game.White) + msgLeft + strings.Repeat(" ", max(0, 20-len(msgLeft)-len(msgRight))) + msgRight + string(game.Reset)
			}
		}
	}
	return frame
}
