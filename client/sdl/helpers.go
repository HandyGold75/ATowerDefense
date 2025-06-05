package clsdl

import (
	"ATowerDefense/game"
	"os"
	"strings"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

func newTextures(r *sdl.Renderer) (textures, error) {
	execPath, err := os.Executable()
	if err != nil {
		return textures{}, err
	}
	fileSplit := strings.Split(strings.ReplaceAll(execPath, "\\", "/"), "/")
	execPath = strings.Join(fileSplit[:len(fileSplit)-1], "/")

	loadTexture := func(file string) (*sdl.Texture, error) {
		srf, err := img.LoadPNGRW(sdl.RWFromFile(file, "rb"))
		if err != nil {
			return nil, err
		}
		defer srf.Free()
		txr, err := r.CreateTextureFromSurface(srf)
		if err != nil {
			return nil, err
		}
		return txr, nil
	}

	txrText, err := loadTexture(execPath + "/client/assets/Text.png")
	if err != nil {
		return textures{}, err
	}

	txrUI, err := loadTexture(execPath + "/client/assets/UI.png")
	if err != nil {
		return textures{}, err
	}

	txrObstacles, err := loadTexture(execPath + "/client/assets/Obstacles.png")
	if err != nil {
		return textures{}, err
	}

	txrRoads, err := loadTexture(execPath + "/client/assets/Roads.png")
	if err != nil {
		return textures{}, err
	}

	txrTowers, err := loadTexture(execPath + "/client/assets/Towers.png")
	if err != nil {
		return textures{}, err
	}

	txrEnemies, err := loadTexture(execPath + "/client/assets/Enemies.png")
	if err != nil {
		return textures{}, err
	}

	return textures{
		text:      txrText,
		ui:        txrUI,
		obstacles: txrObstacles,
		roads:     txrRoads,
		towers:    txrTowers,
		enemies:   txrEnemies,
	}, nil
}

func (cl *SDL) newRect(x, y, w, h int32) sdl.Rect {
	return sdl.Rect{X: x * cl.tileW, Y: y * cl.tileH, W: w * cl.tileW, H: h * cl.tileH}
}

func (cl *SDL) newRectP(x, y, w, h int32) *sdl.Rect {
	return &sdl.Rect{X: x * cl.tileW, Y: y * cl.tileH, W: w * cl.tileW, H: h * cl.tileH}
}

func (cl *SDL) srcObstacle(obj *game.ObstacleObj) sdl.Rect {
	switch obj.Name {
	case "lake":
		return cl.newRect(0, 0, 1, 1)
	case "sea":
		return cl.newRect(1, 0, 1, 1)
	case "sand":
		return cl.newRect(2, 0, 1, 1)
	case "hills":
		return cl.newRect(0, 1, 1, 1)
	case "tree":
		return cl.newRect(1, 1, 1, 1)
	case "brick":
		return cl.newRect(2, 1, 1, 1)
	}
	return cl.newRect(0, 0, 0, 0)
}

func (cl *SDL) srcRoad(obj *game.RoadObj) sdl.Rect {
	if obj.Index == 0 {
		switch obj.DirExit {
		case "up":
			return cl.newRect(0, 2, 1, 1)
		case "right":
			return cl.newRect(1, 2, 1, 1)
		case "down":
			return cl.newRect(2, 2, 1, 1)
		case "left":
			return cl.newRect(3, 2, 1, 1)
		}
	} else if obj.Index == len(cl.game.GS.Roads)-1 {
		switch obj.DirEntrance {
		case "up":
			return cl.newRect(0, 3, 1, 1)
		case "right":
			return cl.newRect(1, 3, 1, 1)
		case "down":
			return cl.newRect(2, 3, 1, 1)
		case "left":
			return cl.newRect(3, 3, 1, 1)
		}
	} else {
		switch obj.DirEntrance + ";" + obj.DirExit {
		case "up;down", "down;up":
			return cl.newRect(0, 0, 1, 1)
		case "left;right", "right;left":
			return cl.newRect(1, 0, 1, 1)
		case "up;right", "right;up":
			return cl.newRect(0, 1, 1, 1)
		case "right;down", "down;right":
			return cl.newRect(1, 1, 1, 1)
		case "down;left", "left;down":
			return cl.newRect(2, 1, 1, 1)
		case "left;up", "up;left":
			return cl.newRect(3, 1, 1, 1)
		}
	}
	return cl.newRect(0, 0, 0, 0)
}

func (cl *SDL) srcTower(obj *game.TowerObj) sdl.Rect {
	switch obj.Name {
	case "Basic":
		return cl.newRect(int32((obj.Rotation/360)*16), 0, 1, 1)
	case "LongRange":
		return cl.newRect(int32((obj.Rotation/360)*16), 1, 1, 1)
	case "Fast":
		return cl.newRect(int32((obj.Rotation/360)*16), 2, 1, 1)
	case "Strong":
		return cl.newRect(int32((obj.Rotation/360)*16), 3, 1, 1)
	}
	return cl.newRect(int32((obj.Rotation/360)*16), 0, 1, 1)
}

func (cl *SDL) srcEnemy(obj *game.EnemyObj, dst *sdl.Rect) sdl.Rect {
	road := cl.game.GS.Roads[min(int(obj.Progress), len(cl.game.GS.Roads)-1)]

	offset := (obj.Progress - float64(int(obj.Progress))) / 2
	switch road.DirEntrance {
	case "up":
		dst.Y += int32(float64(cl.tileH) * offset)
	case "right":
		dst.X -= int32(float64(cl.tileW) * offset)
	case "down":
		dst.Y -= int32(float64(cl.tileH) * offset)
	case "left":
		dst.X += int32(float64(cl.tileW) * offset)
	}

	switch road.DirExit {
	case "up":
		dst.Y -= int32(float64(cl.tileH) * offset)
	case "right":
		dst.X += int32(float64(cl.tileW) * offset)
	case "down":
		dst.Y += int32(float64(cl.tileH) * offset)
	case "left":
		dst.X -= int32(float64(cl.tileW) * offset)
	}

	switch road.DirEntrance + ";" + road.DirExit {
	case "up;down":
		return cl.newRect(0, 0, 1, 1)
	case "up;left", "right;down":
		return cl.newRect(1, 0, 1, 1)
	case "right;left":
		return cl.newRect(2, 0, 1, 1)
	case "right;up", "down;left":
		return cl.newRect(3, 0, 1, 1)
	case "down;up":
		return cl.newRect(4, 0, 1, 1)
	case "down;right", "left;up":
		return cl.newRect(5, 0, 1, 1)
	case "left;right":
		return cl.newRect(6, 0, 1, 1)
	case "left;down", "up;right":
		return cl.newRect(7, 0, 1, 1)
	}

	return cl.newRect(0, 0, 0, 0)
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
