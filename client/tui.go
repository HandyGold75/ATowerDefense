package client

import (
	"ATowerDefense/game"
	"fmt"
	"os"
	"strconv"
	"strings"

	"golang.org/x/term"
)

func drawField(gm *game.Game) {
	fmt.Print("\033[2;0H")
	for y := range min(gm.GC.FieldHeight, maxHeight) {
		if y != 0 {
			fmt.Print("\r\n")
		}
		if y+viewOffsetY < 0 || y+viewOffsetY >= gm.GC.FieldHeight {
			fmt.Print(strings.Repeat(string(game.BGBrightBlack+"  "+game.Reset), min(gm.GC.FieldWidth, maxWidth)))
			continue
		}
		for x := range min(gm.GC.FieldWidth, maxWidth) {
			if x+viewOffsetX < 0 || x+viewOffsetX >= gm.GC.FieldWidth {
				fmt.Print(game.BGBrightBlack + "  " + game.Reset)
			} else if x+viewOffsetX == selectedX && y+viewOffsetY == selectedY {
				fmt.Print(game.BGGreen + game.Black + "" + game.Reset)
			} else if obj := gm.GetCollisions(x+viewOffsetX, y+viewOffsetY); len(obj) > 0 {
				switch obj[len(obj)-1].Type() {
				case "Obstacle":
					fmt.Print(obj[len(obj)-1].Color() + "" + game.Reset)

				case "Road":
					if obj[len(obj)-1].(*game.RoadObj).Index == 0 {
						fmt.Print(obj[len(obj)-1].Color() + game.BrightBlack + " 󰮢" + game.Reset)
						continue
					} else if obj[len(obj)-1].(*game.RoadObj).Index == len(gm.GS.Roads)-1 {
						fmt.Print(obj[len(obj)-1].Color() + game.BrightBlack + " 󰄚" + game.Reset)
						continue
					}

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
					roadX, roadY := gm.GS.Roads[0].Cord()
					enemyX, enemyY := obj[len(obj)-1].Cord()
					if enemyX == roadX && enemyY == roadY {
						fmt.Print(obj[len(obj)-1].Color() + game.BrightBlack + " 󰮢" + game.Reset)
						continue
					}
					fmt.Print(obj[len(obj)-1].Color() + " " + game.Reset)

				default:
					fmt.Print(obj[len(obj)-1].Color() + "??" + game.Reset)
				}
			} else {
				fmt.Print(game.Green + "██" + game.Reset)
			}
		}
	}
}

func drawUI(gm *game.Game) {
	state := gm.GS.Phase
	if gm.GS.State == "paused" {
		state += " [p]"
	}
	phase := "R:" + strconv.Itoa(gm.GS.Round+1)
	if gm.GS.Phase == "defending" {
		phase = "E:" + strconv.Itoa(len(gm.GS.Enemies))
	}
	msgLen := len(state) + len(phase) + 1
	msgLeft := fmt.Sprintf(string(game.BrightWhite)+"%v %v", state, phase)

	lag := strconv.FormatInt(lagTracker.Milliseconds(), 10)
	if lagTracker >= tickDelay {
		msgLen -= 4
		lag = string(game.Red) + lag
	}
	msgLen += len(lag) + len(strconv.Itoa(gm.Players[pid].Coins)) + len(strconv.Itoa(gm.GS.Health)) + 2
	msgRight := fmt.Sprintf(string(game.White)+"%v "+
		string(game.BrightYellow)+"%v "+
		string(game.BrightRed)+"%v", lag, gm.Players[pid].Coins, gm.GS.Health)

	fmt.Printf("\033[0;0H"+string(game.BGBrightBlack)+"%v"+strings.Repeat(" ", max(1, min(gm.GC.FieldWidth*2, maxWidth*2)-msgLen))+"%v"+string(game.Reset), msgLeft, msgRight)

	if maxWidth > gm.GC.FieldWidth+10 && maxHeight+1 >= len(game.Towers) {
		for i, tower := range game.Towers {
			fmt.Print("\033[" + strconv.Itoa(i+1) + ";" + strconv.Itoa((gm.GC.FieldWidth*2)+1) + "H")
			msgLeft := tower.Name
			msgRight := "(" + strconv.Itoa(tower.Cost) + ")"
			if i == selectedTower {
				fmt.Print(string(game.BGWhite+game.Black) + msgLeft + strings.Repeat(" ", max(0, 20-len(msgLeft)-len(msgRight))) + msgRight + string(game.Reset))
			} else {
				fmt.Print(string(game.BGBlack+game.White) + msgLeft + strings.Repeat(" ", max(0, 20-len(msgLeft)-len(msgRight))) + msgRight + string(game.Reset))
			}
		}
	}
}

func drawTui(gm *game.Game) error {
	fmt.Print("\033[2J")
	mw, mh, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	maxWidth, maxHeight = int(mw/2), mh-1

	drawField(gm)
	drawUI(gm)

	fmt.Print("\033[" + strconv.Itoa(maxHeight) + ";" + strconv.Itoa(maxWidth*2) + "H")
	return nil
}
