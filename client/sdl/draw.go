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
)

func (cl *SDL) newRect(v vector) sdl.Rect {
	return sdl.Rect{X: v.x * cl.tileW, Y: v.y * cl.tileH, W: cl.tileW, H: cl.tileH}
}

func (cl *SDL) newRectP(v vector) *sdl.Rect {
	return &sdl.Rect{X: v.x * cl.tileW, Y: v.y * cl.tileH, W: cl.tileW, H: cl.tileH}
}

func (cl *SDL) srcEnemy(obj *game.EnemyObj, dst *sdl.Rect) *sdl.Rect {
	road := cl.GM.GS.Roads[min(int(obj.Progress), len(cl.GM.GS.Roads)-1)]

	// 0.0 - 0.5; lower makes the rotate anamation longer
	rotateOffset := float64(1) / 3

	progdec := (obj.Progress - float64(int(obj.Progress)))
	if obj.Progress < 1 {
		progdec = (progdec * rotateOffset) + (1 - rotateOffset)
	} else if int(obj.Progress) >= len(cl.GM.GS.Roads)-1 {
		progdec = (progdec * rotateOffset)
	}

	switch {
	case progdec <= rotateOffset:
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
		return cl.newRectP(TEXTURE_ENEMIES[road.DirEntrance+";end"])

	case progdec >= 1-rotateOffset:
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
		return cl.newRectP(TEXTURE_ENEMIES["start;"+road.DirExit])

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
		return cl.newRectP(TEXTURE_ENEMIES[road.DirEntrance+";"+road.DirExit])
	}
}

func (cl *SDL) renderString(str string, x, y int32) error {
	sources := []sdl.Rect{}
	for _, char := range str {
		sources = append(sources, cl.newRect(TEXTURE_TEXT[char]))
	}

	for i, src := range sources {
		if err := cl.renderer.Copy(cl.Textures.text, &src, &sdl.Rect{X: x + ((cl.tileW / 2) * int32(i)), Y: y, W: cl.tileW, H: cl.tileH}); err != nil {
			return err
		}
	}
	return nil
}

func (cl *SDL) drawField() error {
	if err := cl.renderer.SetDrawColor(0, 255, 0, 255); err != nil {
		return err
	}

	for y := range cl.GM.GC.FieldHeight {
		for x := range cl.GM.GC.FieldWidth {
			dst := cl.newRect(vector{int32(x + cl.viewOffsetX), int32(y + cl.viewOffsetY)})
			bgSrc, ok := backgroundCache[x][y]
			if !ok {
				bgSrc := cl.newRect(TEXTURE_BACKGROUND[rand.Int32N(6)])
				if _, ok := backgroundCache[x]; !ok {
					backgroundCache[x] = map[int]sdl.Rect{}
				}
				backgroundCache[x][y] = bgSrc
			}
			if err := cl.renderer.Copy(cl.Textures.environment, &bgSrc, &dst); err != nil {
				return err
			}

			for _, obj := range cl.GM.GetCollisions(x, y) {
				switch obj := obj.(type) {
				case *game.ObstacleObj:
					src, ok := obstacleCache[obj.UID]
					if !ok {
						src = cl.newRect(TEXTURE_BACKGROUND[rand.Int32N(6)])
						obstacleCache[obj.UID] = src
					}

					if err := cl.renderer.Copy(cl.Textures.environment, &src, &dst); err != nil {
						return err
					}

				case *game.RoadObj:
					src := cl.newRect(TEXTURE_ROADS[obj.DirEntrance+";"+obj.DirExit])
					if err := cl.renderer.Copy(cl.Textures.roads, &src, &dst); err != nil {
						return err
					}

				case *game.TowerObj:
					src := cl.newRect(TEXTURE_TOWERS[obj.Name][min(int32((obj.Rotation/360)*16), 15)])
					if err := cl.renderer.Copy(cl.Textures.towers, &src, &dst); err != nil {
						return err
					}
					dstOffset := dst
					dstOffset.Y -= int32(float64(cl.tileH) * 0.75)
					if err := cl.renderer.Copy(cl.Textures.ui, cl.newRectP(vector{int32(min(obj.ReloadProgress, 1) * 9), 2}), &dstOffset); err != nil {
						return err
					}

				default:
					continue
				}
			}
		}
	}

	for y := range cl.GM.GC.FieldHeight {
		for x := range cl.GM.GC.FieldWidth {
			dst := cl.newRect(vector{int32(x + cl.viewOffsetX), int32(y + cl.viewOffsetY)})

			for _, obj := range cl.GM.GetCollisionEnemies(x, y) {
				if obj.Progress == 0.0 {
					continue
				}

				dstOffset := dst
				if err := cl.renderer.Copy(cl.Textures.enemies, cl.srcEnemy(obj, &dstOffset), &dstOffset); err != nil {
					return err
				}
				dstOffset.Y -= int32(float64(cl.tileH) * 0.75)
				if err := cl.renderer.Copy(cl.Textures.ui, cl.newRectP(vector{int32((float64(obj.Health) / float64(obj.StartHealth)) * 9), 1}), &dstOffset); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (cl *SDL) drawUI(processTime time.Duration) error {
	if cl.GM.GS.Phase == "building" {
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

	if err := cl.renderer.Copy(cl.Textures.ui, cl.newRectP(vector{0, 0}), cl.newRectP(vector{int32(cl.selectedX + cl.viewOffsetX), int32(cl.selectedY + cl.viewOffsetY)})); err != nil {
		return err
	}

	phase := cl.GM.GS.Phase + " R:" + strconv.Itoa(cl.GM.GS.Round)
	if cl.GM.GS.Phase == "defending" {
		phase += " E:" + strconv.Itoa(len(cl.GM.GS.Enemies))
	}

	if err := cl.renderString(phase, 0, 0); err != nil {
		return err
	}

	// if processTime >= cl.GM.GC.TickDelay {
	// }
	stats := fmt.Sprintf("%v %v %v %v", cl.GM.GC.GameSpeed, processTime.Milliseconds(), cl.GM.Players[cl.pid].Coins, cl.GM.GS.Health)
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

	if cl.GM.GS.State == "paused" {
		msg := "Paused"
		if err := cl.renderString(msg, (cl.windowW/2)-(cl.tileW/2)-((cl.tileW/2)*int32(len(msg)/2)), (cl.windowH/2)-(cl.tileH/2)); err != nil {
			return err
		}
	}

	if cl.GM.GS.Phase == "lost" {
		msg := "Game Over"
		if err := cl.renderString(msg, (cl.windowW/2)-(cl.tileW/2)-((cl.tileW/2)*int32(len(msg)/2)), (cl.windowH/2)-(cl.tileH/2)); err != nil {
			return err
		}
	}

	return nil
}
