package clsdl

import (
	"ATowerDefense/game"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

var (
	backgroundCache = map[int]map[int]sdl.Rect{}
	obstacleCache   = map[int]sdl.Rect{}

	// 0.0 - 0.5; lower makes the rotate anamation longer
	rotateAnimationOffset = float64(1) / 3
)

func (cl *clSDL) newRect(v vector) sdl.Rect {
	return sdl.Rect{X: v.x * cl.tileW, Y: v.y * cl.tileH, W: cl.tileW, H: cl.tileH}
}

func (cl *clSDL) newRectP(v vector) *sdl.Rect {
	return &sdl.Rect{X: v.x * cl.tileW, Y: v.y * cl.tileH, W: cl.tileW, H: cl.tileH}
}

func (cl *clSDL) renderString(str string, x, y int32) error {
	sources := []sdl.Rect{}
	for _, char := range str {
		sources = append(sources, cl.newRect(TEXTURE_TEXT[char]))
	}

	for i, src := range sources {
		if err := cl.renderer.Copy(cl.textures.text, &src, &sdl.Rect{X: x + ((cl.tileW / 2) * int32(i)), Y: y, W: cl.tileW, H: cl.tileH}); err != nil {
			return err
		}
	}
	return nil
}

func (cl *clSDL) drawField() error {
	for y := range cl.gm.GC.FieldHeight {
		for x := range cl.gm.GC.FieldWidth {
			dst := cl.newRect(vector{int32(x + cl.viewOffsetX), int32(y + cl.viewOffsetY)})
			src, ok := backgroundCache[x][y]
			if !ok {
				src = cl.newRect(TEXTURE_BACKGROUND[rand.Int32N(6)])
				if _, ok := backgroundCache[x]; !ok {
					backgroundCache[x] = map[int]sdl.Rect{}
				}
				backgroundCache[x][y] = src
			}
			if err := cl.renderer.Copy(cl.textures.environment, &src, &dst); err != nil {
				return err
			}
		}
	}

	for _, road := range cl.gm.GS.Roads {
		x, y := road.Cord()
		dst := cl.newRect(vector{int32(x + cl.viewOffsetX), int32(y + cl.viewOffsetY)})
		src := cl.newRect(TEXTURE_ROADS[road.DirEntrance+";"+road.DirExit])
		if err := cl.renderer.Copy(cl.textures.roads, &src, &dst); err != nil {
			return err
		}
	}

	for _, tower := range cl.gm.GS.Towers {
		x, y := tower.Cord()
		dst := cl.newRect(vector{int32(x + cl.viewOffsetX), int32(y + cl.viewOffsetY)})
		src := cl.newRect(TEXTURE_TOWERS[tower.Name][min(int32((tower.Rotation/360)*16), 15)])
		if err := cl.renderer.Copy(cl.textures.towers, &src, &dst); err != nil {
			return err
		}
		dst.Y -= int32(float64(cl.tileH) * 0.75)
		if err := cl.renderer.Copy(cl.textures.ui, cl.newRectP(vector{int32(min(tower.ReloadProgress, 1) * 9), 2}), &dst); err != nil {
			return err
		}
	}

	for _, obstacle := range cl.gm.GS.Obstacles {
		x, y := obstacle.Cord()
		dst := cl.newRect(vector{int32(x + cl.viewOffsetX), int32(y + cl.viewOffsetY)})
		src, ok := obstacleCache[obstacle.UID]
		if !ok {
			src = cl.newRect(TEXTURE_OBSTACLES[rand.Int32N(6)])
			obstacleCache[obstacle.UID] = src
		}
		if err := cl.renderer.Copy(cl.textures.environment, &src, &dst); err != nil {
			return err
		}
	}

	for _, enemy := range cl.gm.GS.Enemies {
		if enemy.Progress == 0.0 {
			continue
		}

		x, y := enemy.Cord()
		dst := cl.newRect(vector{int32(x + cl.viewOffsetX), int32(y + cl.viewOffsetY)})
		road := cl.gm.GS.Roads[min(int(enemy.Progress), len(cl.gm.GS.Roads)-1)]
		src := cl.newRect(TEXTURE_ENEMIES[road.DirEntrance+";"+road.DirExit])

		progdec := (enemy.Progress - float64(int(enemy.Progress)))
		if enemy.Progress < 1 {
			progdec = (progdec * rotateAnimationOffset) + (1 - rotateAnimationOffset)
		} else if int(enemy.Progress) >= len(cl.gm.GS.Roads)-1 {
			progdec = (progdec * rotateAnimationOffset)
		}

		switch {
		case progdec <= rotateAnimationOffset:
			switch road.DirEntrance {
			case "up":
				dst.Y -= int32(float64(cl.tileH) * (0.5 - progdec))
			case "right":
				dst.X += int32(float64(cl.tileW) * (0.5 - progdec))
			case "down":
				dst.Y += int32(float64(cl.tileH) * (0.5 - progdec))
			case "left":
				dst.X -= int32(float64(cl.tileW) * (0.5 - progdec))
			}
			src = cl.newRect(TEXTURE_ENEMIES[road.DirEntrance+";end"])

		case progdec >= 1-rotateAnimationOffset:
			switch road.DirExit {
			case "up":
				dst.Y -= int32(float64(cl.tileH) * (progdec - 0.5))
			case "right":
				dst.X += int32(float64(cl.tileW) * (progdec - 0.5))
			case "down":
				dst.Y += int32(float64(cl.tileH) * (progdec - 0.5))
			case "left":
				dst.X -= int32(float64(cl.tileW) * (progdec - 0.5))
			}
			src = cl.newRect(TEXTURE_ENEMIES["start;"+road.DirExit])

		default:
			switch road.DirEntrance {
			case "up":
				dst.Y -= int32(float64(cl.tileH) * (0.25 - (progdec / 2)))
			case "right":
				dst.X += int32(float64(cl.tileW) * (0.25 - (progdec / 2)))
			case "down":
				dst.Y += int32(float64(cl.tileH) * (0.25 - (progdec / 2)))
			case "left":
				dst.X -= int32(float64(cl.tileW) * (0.25 - (progdec / 2)))
			}
			switch road.DirExit {
			case "up":
				dst.Y -= int32(float64(cl.tileH) * ((progdec / 2) - 0.25))
			case "right":
				dst.X += int32(float64(cl.tileW) * ((progdec / 2) - 0.25))
			case "down":
				dst.Y += int32(float64(cl.tileH) * ((progdec / 2) - 0.25))
			case "left":
				dst.X -= int32(float64(cl.tileW) * ((progdec / 2) - 0.25))
			}
		}

		if err := cl.renderer.Copy(cl.textures.enemies, &src, &dst); err != nil {
			return err
		}
		dst.Y -= int32(float64(cl.tileH) * 0.75)
		if err := cl.renderer.Copy(cl.textures.ui, cl.newRectP(vector{int32((float64(enemy.Health) / float64(enemy.StartHealth)) * 9), 1}), &dst); err != nil {
			return err
		}
	}

	return nil
}

func (cl *clSDL) drawUI(processTime time.Duration) error {
	if cl.gm.GS.Phase == "building" {
		if err := cl.renderer.SetDrawColor(255, 0, 0, 85); err != nil {
			return err
		}
		r := game.Towers[cl.selectedTower].Range
		dst := cl.newRectP(vector{int32(cl.selectedX + cl.viewOffsetX - r), int32(cl.selectedY + cl.viewOffsetY - r)})
		dst.W, dst.H = int32((r*2)+1)*cl.tileW, int32((r*2)+1)*cl.tileH
		if err := cl.renderer.FillRect(dst); err != nil {
			return err
		}
	}

	if err := cl.renderer.Copy(cl.textures.ui, cl.newRectP(vector{0, 0}), cl.newRectP(vector{int32(cl.selectedX + cl.viewOffsetX), int32(cl.selectedY + cl.viewOffsetY)})); err != nil {
		return err
	}

	phase := cl.gm.GS.Phase + " R:" + strconv.Itoa(cl.gm.GS.Round)
	if cl.gm.GS.Phase == "defending" {
		phase += " E:" + strconv.Itoa(len(cl.gm.GS.Enemies))
	}

	if err := cl.renderString(phase, 0, 0); err != nil {
		return err
	}

	// if processTime >= cl.gm.GC.TickDelay {
	// }
	stats := fmt.Sprintf("%v %v %v %v", cl.gm.GC.GameSpeed, processTime.Microseconds(), cl.gm.Players[cl.pid].Coins, cl.gm.GS.Health)
	stats = strings.Repeat(" ", int(cl.windowW/32)-len(stats)-1) + stats

	if err := cl.renderString(stats, 0, 0); err != nil {
		return err
	}

	for i, tower := range game.Towers {
		if i == cl.selectedTower {
			if err := cl.renderString(tower.Name+" <", 0, (cl.windowH-(cl.tileH*int32(len(game.Towers))))+(cl.tileH*int32(i))); err != nil {
				return err
			}
			continue
		}
		if err := cl.renderString(tower.Name, 0, (cl.windowH-(cl.tileH*int32(len(game.Towers))))+(cl.tileH*int32(i))); err != nil {
			return err
		}
	}

	if cl.gm.GS.State == "paused" {
		msg := "Paused"
		if err := cl.renderString(msg, (cl.windowW/2)-(cl.tileW/2)-((cl.tileW/2)*int32(len(msg)/2)), (cl.windowH/2)-(cl.tileH/2)); err != nil {
			return err
		}
	}

	if cl.gm.GS.Phase == "lost" {
		msg := "Game Over"
		if err := cl.renderString(msg, (cl.windowW/2)-(cl.tileW/2)-((cl.tileW/2)*int32(len(msg)/2)), (cl.windowH/2)-(cl.tileH/2)); err != nil {
			return err
		}
	}

	return nil
}
