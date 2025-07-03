package clsdl

import (
	"ATowerDefense/game"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

func (cl *SDL) newRect(x, y, w, h int32) sdl.Rect {
	return sdl.Rect{X: x * cl.tileW, Y: y * cl.tileH, W: w * cl.tileW, H: h * cl.tileH}
}

func (cl *SDL) newRectP(x, y, w, h int32) *sdl.Rect {
	return &sdl.Rect{X: x * cl.tileW, Y: y * cl.tileH, W: w * cl.tileW, H: h * cl.tileH}
}

func (cl *SDL) srcObstacle(obj *game.ObstacleObj) *sdl.Rect {
	switch obj.Name {
	case "lake":
		return cl.newRectP(0, 0, 1, 1)
	case "sea":
		return cl.newRectP(1, 0, 1, 1)
	case "sand":
		return cl.newRectP(2, 0, 1, 1)
	case "hills":
		return cl.newRectP(0, 1, 1, 1)
	case "tree":
		return cl.newRectP(1, 1, 1, 1)
	case "brick":
		return cl.newRectP(2, 1, 1, 1)
	}
	return cl.newRectP(0, 0, 0, 0)
}

func (cl *SDL) srcRoad(obj *game.RoadObj) *sdl.Rect {
	if obj.Index == 0 {
		switch obj.DirExit {
		case "up":
			return cl.newRectP(0, 2, 1, 1)
		case "right":
			return cl.newRectP(1, 2, 1, 1)
		case "down":
			return cl.newRectP(2, 2, 1, 1)
		case "left":
			return cl.newRectP(3, 2, 1, 1)
		}
	} else if obj.Index == len(cl.game.GS.Roads)-1 {
		switch obj.DirEntrance {
		case "up":
			return cl.newRectP(0, 3, 1, 1)
		case "right":
			return cl.newRectP(1, 3, 1, 1)
		case "down":
			return cl.newRectP(2, 3, 1, 1)
		case "left":
			return cl.newRectP(3, 3, 1, 1)
		}
	} else {
		switch obj.DirEntrance + ";" + obj.DirExit {
		case "up;down", "down;up":
			return cl.newRectP(0, 0, 1, 1)
		case "left;right", "right;left":
			return cl.newRectP(1, 0, 1, 1)
		case "up;right", "right;up":
			return cl.newRectP(0, 1, 1, 1)
		case "right;down", "down;right":
			return cl.newRectP(1, 1, 1, 1)
		case "down;left", "left;down":
			return cl.newRectP(2, 1, 1, 1)
		case "left;up", "up;left":
			return cl.newRectP(3, 1, 1, 1)
		}
	}
	return cl.newRectP(0, 0, 0, 0)
}

func (cl *SDL) srcTower(obj *game.TowerObj) *sdl.Rect {
	switch obj.Name {
	case "Basic":
		return cl.newRectP(int32((obj.Rotation/360)*16), 0, 1, 1)
	case "LongRange":
		return cl.newRectP(int32((obj.Rotation/360)*16), 1, 1, 1)
	case "Fast":
		return cl.newRectP(int32((obj.Rotation/360)*16), 2, 1, 1)
	case "Strong":
		return cl.newRectP(int32((obj.Rotation/360)*16), 3, 1, 1)
	}
	return cl.newRectP(int32((obj.Rotation/360)*16), 0, 1, 1)
}

func (cl *SDL) srcEnemy(obj *game.EnemyObj, dst *sdl.Rect) *sdl.Rect {
	road := cl.game.GS.Roads[min(int(obj.Progress), len(cl.game.GS.Roads)-1)]

	offset := (obj.Progress - float64(int(obj.Progress)))
	if obj.Progress < 1 {
		offset = (offset / 2) + 0.5
	} else if int(obj.Progress) >= len(cl.game.GS.Roads)-1 {
		offset = (offset / 2)
	}

	switch road.DirEntrance {
	case "up":
		dst.Y += int32(float64(cl.tileH) * (min(offset, 0.5) - 0.5))
	case "right":
		dst.X -= int32(float64(cl.tileW) * (min(offset, 0.5) - 0.5))
	case "down":
		dst.Y -= int32(float64(cl.tileH) * (min(offset, 0.5) - 0.5))
	case "left":
		dst.X += int32(float64(cl.tileW) * (min(offset, 0.5) - 0.5))
	}

	switch road.DirExit {
	case "up":
		dst.Y -= int32(float64(cl.tileH) * (max(offset, 0.5) - 0.5))
	case "right":
		dst.X += int32(float64(cl.tileW) * (max(offset, 0.5) - 0.5))
	case "down":
		dst.Y += int32(float64(cl.tileH) * (max(offset, 0.5) - 0.5))
	case "left":
		dst.X -= int32(float64(cl.tileW) * (max(offset, 0.5) - 0.5))
	}

	switch road.DirEntrance + ";" + road.DirExit {
	case "up;down":
		return cl.newRectP(0, 0, 1, 1)
	case "up;left", "right;down":
		return cl.newRectP(1, 0, 1, 1)
	case "right;left":
		return cl.newRectP(2, 0, 1, 1)
	case "right;up", "down;left":
		return cl.newRectP(3, 0, 1, 1)
	case "down;up":
		return cl.newRectP(4, 0, 1, 1)
	case "down;right", "left;up":
		return cl.newRectP(5, 0, 1, 1)
	case "left;right":
		return cl.newRectP(6, 0, 1, 1)
	case "left;down", "up;right":
		return cl.newRectP(7, 0, 1, 1)
	}

	return cl.newRectP(0, 0, 0, 0)
}

func (cl *SDL) renderString(str string, x, y int32) error {
	sources := []sdl.Rect{}
	for _, char := range str {
		switch char {
		case ' ':
			sources = append(sources, sdl.Rect{X: cl.tileW * -1, Y: cl.tileH * -1, W: cl.tileW, H: cl.tileH})

		case '0':
			sources = append(sources, sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH})
		case '1':
			sources = append(sources, sdl.Rect{X: cl.tileW * 1, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH})
		case '2':
			sources = append(sources, sdl.Rect{X: cl.tileW * 2, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH})
		case '3':
			sources = append(sources, sdl.Rect{X: cl.tileW * 3, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH})
		case '4':
			sources = append(sources, sdl.Rect{X: cl.tileW * 4, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH})
		case '5':
			sources = append(sources, sdl.Rect{X: cl.tileW * 5, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH})
		case '6':
			sources = append(sources, sdl.Rect{X: cl.tileW * 6, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH})
		case '7':
			sources = append(sources, sdl.Rect{X: cl.tileW * 7, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH})
		case '8':
			sources = append(sources, sdl.Rect{X: cl.tileW * 8, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH})
		case '9':
			sources = append(sources, sdl.Rect{X: cl.tileW * 9, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH})

		case 'a':
			sources = append(sources, sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'b':
			sources = append(sources, sdl.Rect{X: cl.tileW * 1, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'c':
			sources = append(sources, sdl.Rect{X: cl.tileW * 2, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'd':
			sources = append(sources, sdl.Rect{X: cl.tileW * 3, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'e':
			sources = append(sources, sdl.Rect{X: cl.tileW * 4, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'f':
			sources = append(sources, sdl.Rect{X: cl.tileW * 5, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'g':
			sources = append(sources, sdl.Rect{X: cl.tileW * 6, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'h':
			sources = append(sources, sdl.Rect{X: cl.tileW * 7, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'i':
			sources = append(sources, sdl.Rect{X: cl.tileW * 8, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'j':
			sources = append(sources, sdl.Rect{X: cl.tileW * 9, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'k':
			sources = append(sources, sdl.Rect{X: cl.tileW * 10, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'l':
			sources = append(sources, sdl.Rect{X: cl.tileW * 11, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'm':
			sources = append(sources, sdl.Rect{X: cl.tileW * 12, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'n':
			sources = append(sources, sdl.Rect{X: cl.tileW * 13, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'o':
			sources = append(sources, sdl.Rect{X: cl.tileW * 14, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'p':
			sources = append(sources, sdl.Rect{X: cl.tileW * 15, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'q':
			sources = append(sources, sdl.Rect{X: cl.tileW * 16, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'r':
			sources = append(sources, sdl.Rect{X: cl.tileW * 17, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 's':
			sources = append(sources, sdl.Rect{X: cl.tileW * 18, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 't':
			sources = append(sources, sdl.Rect{X: cl.tileW * 19, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'u':
			sources = append(sources, sdl.Rect{X: cl.tileW * 20, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'v':
			sources = append(sources, sdl.Rect{X: cl.tileW * 21, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'w':
			sources = append(sources, sdl.Rect{X: cl.tileW * 22, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'x':
			sources = append(sources, sdl.Rect{X: cl.tileW * 23, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'y':
			sources = append(sources, sdl.Rect{X: cl.tileW * 24, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})
		case 'z':
			sources = append(sources, sdl.Rect{X: cl.tileW * 25, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH})

		case 'A':
			sources = append(sources, sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'B':
			sources = append(sources, sdl.Rect{X: cl.tileW * 1, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'C':
			sources = append(sources, sdl.Rect{X: cl.tileW * 2, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'D':
			sources = append(sources, sdl.Rect{X: cl.tileW * 3, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'E':
			sources = append(sources, sdl.Rect{X: cl.tileW * 4, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'F':
			sources = append(sources, sdl.Rect{X: cl.tileW * 5, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'G':
			sources = append(sources, sdl.Rect{X: cl.tileW * 6, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'H':
			sources = append(sources, sdl.Rect{X: cl.tileW * 7, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'I':
			sources = append(sources, sdl.Rect{X: cl.tileW * 8, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'J':
			sources = append(sources, sdl.Rect{X: cl.tileW * 9, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'K':
			sources = append(sources, sdl.Rect{X: cl.tileW * 10, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'L':
			sources = append(sources, sdl.Rect{X: cl.tileW * 11, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'M':
			sources = append(sources, sdl.Rect{X: cl.tileW * 12, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'N':
			sources = append(sources, sdl.Rect{X: cl.tileW * 13, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'O':
			sources = append(sources, sdl.Rect{X: cl.tileW * 14, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'P':
			sources = append(sources, sdl.Rect{X: cl.tileW * 15, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'Q':
			sources = append(sources, sdl.Rect{X: cl.tileW * 16, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'R':
			sources = append(sources, sdl.Rect{X: cl.tileW * 17, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'S':
			sources = append(sources, sdl.Rect{X: cl.tileW * 18, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'T':
			sources = append(sources, sdl.Rect{X: cl.tileW * 19, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'U':
			sources = append(sources, sdl.Rect{X: cl.tileW * 20, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'V':
			sources = append(sources, sdl.Rect{X: cl.tileW * 21, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'W':
			sources = append(sources, sdl.Rect{X: cl.tileW * 22, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'X':
			sources = append(sources, sdl.Rect{X: cl.tileW * 23, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'Y':
			sources = append(sources, sdl.Rect{X: cl.tileW * 24, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})
		case 'Z':
			sources = append(sources, sdl.Rect{X: cl.tileW * 25, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH})

		case '!':
			sources = append(sources, sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '?':
			sources = append(sources, sdl.Rect{X: cl.tileW * 1, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '"':
			sources = append(sources, sdl.Rect{X: cl.tileW * 2, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '#':
			sources = append(sources, sdl.Rect{X: cl.tileW * 3, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '$':
			sources = append(sources, sdl.Rect{X: cl.tileW * 4, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '%':
			sources = append(sources, sdl.Rect{X: cl.tileW * 5, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '&':
			sources = append(sources, sdl.Rect{X: cl.tileW * 6, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '\'':
			sources = append(sources, sdl.Rect{X: cl.tileW * 7, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '(':
			sources = append(sources, sdl.Rect{X: cl.tileW * 8, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case ')':
			sources = append(sources, sdl.Rect{X: cl.tileW * 9, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '*':
			sources = append(sources, sdl.Rect{X: cl.tileW * 10, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '+':
			sources = append(sources, sdl.Rect{X: cl.tileW * 11, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case ',':
			sources = append(sources, sdl.Rect{X: cl.tileW * 12, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '-':
			sources = append(sources, sdl.Rect{X: cl.tileW * 13, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '.':
			sources = append(sources, sdl.Rect{X: cl.tileW * 14, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '/':
			sources = append(sources, sdl.Rect{X: cl.tileW * 15, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case ':':
			sources = append(sources, sdl.Rect{X: cl.tileW * 16, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case ';':
			sources = append(sources, sdl.Rect{X: cl.tileW * 17, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '<':
			sources = append(sources, sdl.Rect{X: cl.tileW * 18, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '=':
			sources = append(sources, sdl.Rect{X: cl.tileW * 19, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '>':
			sources = append(sources, sdl.Rect{X: cl.tileW * 20, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '@':
			sources = append(sources, sdl.Rect{X: cl.tileW * 21, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '[':
			sources = append(sources, sdl.Rect{X: cl.tileW * 22, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '\\':
			sources = append(sources, sdl.Rect{X: cl.tileW * 23, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case ']':
			sources = append(sources, sdl.Rect{X: cl.tileW * 24, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '^':
			sources = append(sources, sdl.Rect{X: cl.tileW * 25, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '_':
			sources = append(sources, sdl.Rect{X: cl.tileW * 26, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '`':
			sources = append(sources, sdl.Rect{X: cl.tileW * 27, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '{':
			sources = append(sources, sdl.Rect{X: cl.tileW * 28, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '|':
			sources = append(sources, sdl.Rect{X: cl.tileW * 29, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '}':
			sources = append(sources, sdl.Rect{X: cl.tileW * 30, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		case '~':
			sources = append(sources, sdl.Rect{X: cl.tileW * 31, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH})
		}
	}

	for i, src := range sources {
		if err := cl.renderer.Copy(cl.textures.text, &src, &sdl.Rect{X: x + ((cl.tileW / 2) * int32(i)), Y: y, W: cl.tileW, H: cl.tileH}); err != nil {
			return err
		}
	}
	return nil
}

func (cl *SDL) drawField() error {
	if err := cl.renderer.SetDrawColor(0, 255, 0, 255); err != nil {
		return err
	}

	for y := range cl.game.GC.FieldHeight {
		for x := range cl.game.GC.FieldWidth {
			dst := cl.newRect(int32(x+cl.viewOffsetX), int32(y+cl.viewOffsetY), 1, 1)
			if err := cl.renderer.FillRect(&dst); err != nil {
				return err
			}

			for _, obj := range cl.game.GetCollisions(x, y) {
				switch obj := obj.(type) {
				case *game.ObstacleObj:
					if err := cl.renderer.Copy(cl.textures.obstacles, cl.srcObstacle(obj), &dst); err != nil {
						return err
					}

				case *game.RoadObj:
					if err := cl.renderer.Copy(cl.textures.roads, cl.srcRoad(obj), &dst); err != nil {
						return err
					}

				case *game.TowerObj:
					if err := cl.renderer.Copy(cl.textures.towers, cl.srcTower(obj), &dst); err != nil {
						return err
					}
					dstOffset := dst
					dstOffset.Y -= int32(float64(cl.tileH) * 0.75)
					if err := cl.renderer.Copy(cl.textures.ui, cl.newRectP(int32(min(obj.ReloadProgress, 1)*9), 2, 1, 1), &dstOffset); err != nil {
						return err
					}

				default:
					continue
				}
			}
		}
	}

	for y := range cl.game.GC.FieldHeight {
		for x := range cl.game.GC.FieldWidth {
			dst := cl.newRect(int32(x+cl.viewOffsetX), int32(y+cl.viewOffsetY), 1, 1)

			for _, obj := range cl.game.GetCollisionEnemies(x, y) {
				if obj.Progress == 0.0 {
					continue
				}

				dstOffset := dst
				if err := cl.renderer.Copy(cl.textures.enemies, cl.srcEnemy(obj, &dstOffset), &dstOffset); err != nil {
					return err
				}
				dstOffset.Y -= int32(float64(cl.tileH) * 0.75)
				if err := cl.renderer.Copy(cl.textures.ui, cl.newRectP(int32((float64(obj.Health)/float64(obj.StartHealth))*9), 1, 1, 1), &dstOffset); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (cl *SDL) drawUI(processTime time.Duration) error {
	if cl.game.GS.Phase == "building" {
		if err := cl.renderer.SetDrawColor(255, 0, 0, 85); err != nil {
			return err
		}
		r := game.Towers[cl.selectedTower].Range
		if err := cl.renderer.FillRect(cl.newRectP(int32(cl.selectedX+cl.viewOffsetX-r), int32(cl.selectedY+cl.viewOffsetY-r), int32((r*2)+1), int32((r*2)+1))); err != nil {
			return err
		}
	}

	if err := cl.renderer.Copy(cl.textures.ui, cl.newRectP(0, 0, 1, 1), cl.newRectP(int32(cl.selectedX+cl.viewOffsetX), int32(cl.selectedY+cl.viewOffsetY), 1, 1)); err != nil {
		return err
	}

	phase := cl.game.GS.Phase + " R:" + strconv.Itoa(cl.game.GS.Round)
	if cl.game.GS.Phase == "defending" {
		phase += " E:" + strconv.Itoa(len(cl.game.GS.Enemies))
	}

	if err := cl.renderString(phase, 0, 0); err != nil {
		return err
	}

	// if processTime >= cl.game.GC.TickDelay {
	// }
	stats := fmt.Sprintf("%v %v %v %v", cl.game.GC.GameSpeed, processTime.Milliseconds(), cl.game.Players[cl.pid].Coins, cl.game.GS.Health)
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

	if cl.game.GS.State == "paused" {
		msg := "Paused"
		if err := cl.renderString(msg, (cl.windowW/2)-(cl.tileW/2)-((cl.tileW/2)*int32(len(msg)/2)), (cl.windowH/2)-(cl.tileH/2)); err != nil {
			return err
		}
	}

	if cl.game.GS.Phase == "lost" {
		msg := "Game Over"
		if err := cl.renderString(msg, (cl.windowW/2)-(cl.tileW/2)-((cl.tileW/2)*int32(len(msg)/2)), (cl.windowH/2)-(cl.tileH/2)); err != nil {
			return err
		}
	}

	return nil
}
