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

	"github.com/HandyGold75/GOLib/tui"
	"golang.org/x/term"
)

type (
	color string

	keybinds struct {
		exit, pause, confirm, delete,
		up, down, right, left,
		panUp, panDown, panRight, panLeft,
		squereBracketLeft, squereBracketRight,
		plus, minus,
		numbers []keybind
	}
	keybind []byte

	clTUI struct {
		gm  *game.Game
		pid int

		oldState *term.State

		selectedX, selectedY,
		viewOffsetX, viewOffsetY,
		selectedTower int

		maxWidth, maxHeight int

		keyBinds keybinds
	}
)

const (
	Reset color = "\033[0m"

	Bold            color = "\033[1m"
	Faint           color = "\033[2m"
	Italic          color = "\033[3m"
	Underline       color = "\033[4m"
	StrikeTrough    color = "\033[9m"
	DubbleUnderline color = "\033[21m"

	Black   color = "\033[30m"
	Red     color = "\033[31m"
	Green   color = "\033[32m"
	Yellow  color = "\033[33m"
	Blue    color = "\033[34m"
	Magenta color = "\033[35m"
	Cyan    color = "\033[36m"
	White   color = "\033[37m"

	BGBlack   color = "\033[40m"
	BGRed     color = "\033[41m"
	BGGreen   color = "\033[42m"
	BGYellow  color = "\033[43m"
	BGBlue    color = "\033[44m"
	BGMagenta color = "\033[45m"
	BGCyan    color = "\033[46m"
	BGWhite   color = "\033[47m"

	BrightBlack   color = "\033[90m"
	BrightRed     color = "\033[91m"
	BrightGreen   color = "\033[92m"
	BrightYellow  color = "\033[93m"
	BrightBlue    color = "\033[94m"
	BrightMagenta color = "\033[95m"
	BrightCyan    color = "\033[96m"
	BrightWhite   color = "\033[97m"

	BGBrightBlack   color = "\033[100m"
	BGBrightRed     color = "\033[101m"
	BGBrightGreen   color = "\033[102m"
	BGBrightYellow  color = "\033[103m"
	BGBrightBlue    color = "\033[104m"
	BGBrightMagenta color = "\033[105m"
	BGBrightCyan    color = "\033[106m"
	BGBrightWhite   color = "\033[107m"
)

func Run(gc game.GameConfig) error {
	cl, err := newTUI(gc)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer cl.stop()
	cl.start()
	return nil
}

func newTUI(gc game.GameConfig) (*clTUI, error) {
	gm := game.NewGame(gc)
	if err := gm.Start(); err != nil {
		return nil, err
	}
	pid := gm.AddPlayer()

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
		return nil, err
	}

	fieldHeight, err := strconv.Atoi(spFieldHeight.Value())
	if err != nil {
		return nil, err
	}
	gc.FieldHeight = fieldHeight
	fieldWidth, err := strconv.Atoi(spFieldWidth.Value())
	if err != nil {
		return nil, err
	}
	gc.FieldWidth = fieldWidth

	gc.IP = mpIP.Value()
	port, err := strconv.ParseUint(mpPort.Value(), 10, 16)
	if err != nil {
		return nil, err
	}
	gc.Port = uint16(port)

	gc.Mode = mode
	gm.GC = gc

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

	return &clTUI{
		gm: gm, pid: pid,
		oldState: state,

		selectedX: 0, selectedY: 0,
		viewOffsetX: 0, viewOffsetY: 0,
		selectedTower: 0,

		maxWidth: int(mw / 2), maxHeight: mh - 1,

		keyBinds: keybinds{
			// ESC, CTRL_C, CTRL_D,
			exit: []keybind{{27, 0, 0}, {3, 0, 0}, {4, 0, 0}},
			// P, Q
			pause: []keybind{{112, 0, 0}, {113, 0, 0}},
			// RETURN
			confirm: []keybind{{13, 0, 0}},
			// BACKSPACE, DEL
			delete: []keybind{{127, 0, 0}, {27, 91, 51}},

			// W, K
			up: []keybind{{119, 0, 0}, {107, 0, 0}},
			// S, J
			down: []keybind{{115, 0, 0}, {106, 0, 0}},
			// D, L
			right: []keybind{{100, 0, 0}, {108, 0, 0}},
			// A, H
			left: []keybind{{97, 0, 0}, {104, 0, 0}},

			// UP
			panUp: []keybind{{27, 91, 65}},
			// DOWN
			panDown: []keybind{{27, 91, 66}},
			// RIGHT,
			panRight: []keybind{{27, 91, 67}},
			// LEFT,
			panLeft: []keybind{{27, 91, 68}},

			// [
			squereBracketLeft: []keybind{{91, 0, 0}},
			// ]
			squereBracketRight: []keybind{{93, 0, 0}},

			// +
			plus: []keybind{{43, 0, 0}},
			// -
			minus: []keybind{{45, 0, 0}},

			// 0, 1, 2, 3, 4, 5, 6, 7, 8, 9
			numbers: []keybind{{48, 0, 0}, {49, 0, 0}, {50, 0, 0}, {51, 0, 0}, {52, 0, 0}, {53, 0, 0}, {54, 0, 0}, {55, 0, 0}, {56, 0, 0}, {57, 0, 0}},
		},
	}, nil
}

func (cl *clTUI) start() {
	go func() {
		defer cl.stop()
		for cl.gm.GS.State != "stopped" {
			if err := cl.input(); err != nil {
				if err == game.Errors.Exit {
					break
				}
				fmt.Println(err)
			}
		}
	}()
	if err := cl.gm.Run(cl.draw); err != nil {
		fmt.Println(err)
	}
}

func (cl *clTUI) stop() {
	if cl.gm.GS.State != "stopped" {
		_ = cl.gm.Stop()
	}

	if cl.oldState != nil {
		_ = term.Restore(int(os.Stdin.Fd()), cl.oldState)
		cl.oldState = nil
	}
	fmt.Print("\r\n")
}

func (cl *clTUI) draw(processTime time.Duration) error {
	mw, mh, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	cl.maxWidth, cl.maxHeight = int(mw/2), mh-1

	fmt.Print("\033[2J" + cl.getField() + cl.getUI(processTime) + "\033[" + strconv.Itoa(cl.maxHeight) + ";" + strconv.Itoa(cl.maxWidth*2) + "H")
	return nil
}

func (cl *clTUI) input() error {
	in := make([]byte, 3)
	if _, err := os.Stdin.Read(in); err != nil {
		return err
	}

	if keyBindContains(cl.keyBinds.exit, in) {
		return game.Errors.Exit
	} else if keyBindContains(cl.keyBinds.pause, in) {
		cl.gm.TogglePause()
		return nil
	} else if keyBindContains(cl.keyBinds.confirm, in) {
		if len(game.Towers) < cl.selectedTower {
			return nil
		}
		if err := cl.gm.PlaceTower(game.Towers[cl.selectedTower].Name, cl.selectedX, cl.selectedY, cl.pid); err != nil {
			if err != game.Errors.InvalidPlacement {
				return err
			}
			if err := cl.gm.DestroyObstacle(cl.selectedX, cl.selectedY, cl.pid); err != nil {
				if err != game.Errors.InvalidPlacement {
					return err
				}
				return cl.gm.DestroyTower(cl.selectedX, cl.selectedY, cl.pid)
			}
		}

	} else if keyBindContains(cl.keyBinds.delete, in) {
		return cl.gm.StartRound()
	} else if keyBindContains(cl.keyBinds.up, in) {
		cl.selectedY = max(cl.selectedY-1, max(0, cl.viewOffsetY))
		return nil
	} else if keyBindContains(cl.keyBinds.down, in) {
		cl.selectedY = min(cl.selectedY+1, min(cl.gm.GC.FieldHeight, min(cl.maxHeight, cl.gm.GC.FieldHeight)+cl.viewOffsetY)-1)
		return nil
	} else if keyBindContains(cl.keyBinds.right, in) {
		cl.selectedX = min(cl.selectedX+1, min(cl.gm.GC.FieldWidth, min(cl.maxWidth, cl.gm.GC.FieldWidth)+cl.viewOffsetX)-1)
		return nil
	} else if keyBindContains(cl.keyBinds.left, in) {
		cl.selectedX = max(cl.selectedX-1, max(0, cl.viewOffsetX))
		return nil

	} else if keyBindContains(cl.keyBinds.panUp, in) {
		cl.viewOffsetY = max(cl.viewOffsetY-1, -5)
		cl.selectedY = max(cl.selectedY-1, max(0, cl.viewOffsetY))
		return nil
	} else if keyBindContains(cl.keyBinds.panDown, in) {
		cl.viewOffsetY = min(cl.viewOffsetY+1, (cl.gm.GC.FieldHeight-min(cl.maxHeight, cl.gm.GC.FieldHeight))+6)
		cl.selectedY = min(cl.selectedY+1, (cl.gm.GC.FieldHeight+min(0, cl.viewOffsetY))-1)
		return nil
	} else if keyBindContains(cl.keyBinds.panRight, in) {
		cl.viewOffsetX = min(cl.viewOffsetX+1, (cl.gm.GC.FieldWidth-min(cl.maxWidth, cl.gm.GC.FieldWidth))+5)
		cl.selectedX = min(cl.selectedX+1, (cl.gm.GC.FieldWidth+min(0, cl.viewOffsetX))-1)
		return nil
	} else if keyBindContains(cl.keyBinds.panLeft, in) {
		cl.viewOffsetX = max(cl.viewOffsetX-1, -5)
		cl.selectedX = max(cl.selectedX-1, max(0, cl.viewOffsetX))
		return nil

	} else if keyBindContains(cl.keyBinds.squereBracketLeft, in) {
		cl.selectedTower = max(cl.selectedTower-1, 0)
		return nil
	} else if keyBindContains(cl.keyBinds.squereBracketRight, in) {
		cl.selectedTower = min(cl.selectedTower+1, len(game.Towers)-1)
		return nil

	} else if keyBindContains(cl.keyBinds.plus, in) {
		cl.gm.GC.GameSpeed = min(cl.gm.GC.GameSpeed+1, 9)
	} else if keyBindContains(cl.keyBinds.minus, in) {
		cl.gm.GC.GameSpeed = max(cl.gm.GC.GameSpeed-1, 0)
	} else if i := keyBindIndex(cl.keyBinds.numbers, in); i >= 0 {
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

func (cl *clTUI) getField() string {
	frame := "\033[2;0H"
	for y := range min(cl.gm.GC.FieldHeight, cl.maxHeight) {
		if y != 0 {
			frame += "\r\n"
		}
		if y+cl.viewOffsetY < 0 || y+cl.viewOffsetY >= cl.gm.GC.FieldHeight {
			frame += strings.Repeat(string(BGBrightBlack+BrightBlack+"  "+Reset), min(cl.gm.GC.FieldWidth, cl.maxWidth))
			continue
		}
		for x := range min(cl.gm.GC.FieldWidth, cl.maxWidth) {
			if x+cl.viewOffsetX < 0 || x+cl.viewOffsetX >= cl.gm.GC.FieldWidth {
				frame += string(BGBrightBlack + BrightBlack + "  " + Reset)
			} else if x+cl.viewOffsetX == cl.selectedX && y+cl.viewOffsetY == cl.selectedY {
				frame += string(BGGreen + Black + "" + Reset)
			} else if objects := cl.gm.GetCollisions(x+cl.viewOffsetX, y+cl.viewOffsetY); len(objects) > 0 {
				switch obj := objects[len(objects)-1].(type) {
				case *game.ObstacleObj:
					frame += string(BGBrightYellow + BrightBlue + "" + Reset)

				case *game.RoadObj:
					if obj.Index == 0 {
						frame += string(BGGreen + White + BrightBlack + " 󰮢" + Reset)
						continue
					} else if obj.Index == len(cl.gm.GS.Roads)-1 {
						frame += string(BGGreen + White + BrightBlack + " 󰄚" + Reset)
						continue
					}

					switch obj.DirExit {
					case "up":
						frame += string(BGGreen + White + " " + Reset)
					case "right":
						frame += string(BGGreen + White + " " + Reset)
					case "down":
						frame += string(BGGreen + White + " " + Reset)
					case "left":
						frame += string(BGGreen + White + " " + Reset)
					default:
						frame += string(BGGreen + White + "?" + Reset)
					}

				case *game.TowerObj:
					frame += string(BGGreen + Black + " 󰚁" + Reset)

				case *game.EnemyObj:
					if obj.Progress < 1 {
						frame += string(BGGreen + Red + BrightBlack + " 󰮢" + Reset)
						continue
					}
					frame += string(BGGreen + Red + " " + Reset)

				default:
					frame += string(BGBrightMagenta + BrightMagenta + "??" + Reset)
				}
			} else {
				frame += string(BGGreen + Green + "██" + Reset)
			}
		}
	}
	return frame
}

func (cl *clTUI) getUI(processTime time.Duration) string {
	phase := cl.gm.GS.Phase
	if cl.gm.GS.State == "paused" {
		phase += " [p]"
	}
	phase += " R:" + strconv.Itoa(cl.gm.GS.Round)
	if cl.gm.GS.Phase == "defending" {
		phase += " E:" + strconv.Itoa(len(cl.gm.GS.Enemies))
	}
	msgLen := len(phase)
	msgLeft := fmt.Sprintf(string(BrightWhite+"%v"), phase)

	lag := strconv.FormatInt(processTime.Milliseconds(), 10)
	if processTime >= cl.gm.GC.TickDelay {
		msgLen -= 4
		lag = string(Red) + lag
	}
	msgLen += len(lag) + len(strconv.Itoa(cl.gm.GC.GameSpeed)) + len(strconv.Itoa(cl.gm.Players[cl.pid].Coins)) + len(strconv.Itoa(cl.gm.GS.Health)) + 3
	msgRight := fmt.Sprintf(string(White+"%v "+White+"%v "+BrightYellow+"%v "+BrightRed+"%v"), lag, cl.gm.GC.GameSpeed, cl.gm.Players[cl.pid].Coins, cl.gm.GS.Health)

	frame := fmt.Sprintf("\033[0;0H"+string(BGBrightBlack)+"%v"+strings.Repeat(" ", max(1, min(cl.gm.GC.FieldWidth*2, cl.maxWidth*2)-msgLen))+"%v"+string(Reset), msgLeft, msgRight)

	if cl.maxWidth > cl.gm.GC.FieldWidth+10 && cl.maxHeight+1 >= len(game.Towers) {
		for i, tower := range game.Towers {
			frame += "\033[" + strconv.Itoa(i+1) + ";" + strconv.Itoa((cl.gm.GC.FieldWidth*2)+1) + "H"
			msgLeft := tower.Name
			msgRight := "(" + strconv.Itoa(tower.Cost) + ")"
			if i == cl.selectedTower {
				frame += string(BGWhite+Black) + msgLeft + strings.Repeat(" ", max(0, 20-len(msgLeft)-len(msgRight))) + msgRight + string(Reset)
			} else {
				frame += string(BGBlack+White) + msgLeft + strings.Repeat(" ", max(0, 20-len(msgLeft)-len(msgRight))) + msgRight + string(Reset)
			}
		}
	}
	return frame
}
