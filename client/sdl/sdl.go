package clsdl

import (
	"ATowerDefense/game"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

type (
	textures struct {
		text      *sdl.Texture
		ui        *sdl.Texture
		obstacles *sdl.Texture
		roads     *sdl.Texture
		towers    *sdl.Texture
		enemies   *sdl.Texture
	}

	SDL struct {
		game *game.Game
		pid  int

		window   *sdl.Window
		renderer *sdl.Renderer

		windowW, windowH,
		tileW, tileH int32

		selectedX, selectedY,
		viewOffsetX, viewOffsetY,
		selectedTower int

		textures textures
	}
)

func NewSDL(gm *game.Game, pid int) (*SDL, error) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return nil, err
	}

	tileSize := int32(64)
	w, err := sdl.CreateWindow("ATowerDefense", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, tileSize*int32(gm.GC.FieldWidth), tileSize*int32(gm.GC.FieldHeight), sdl.WINDOW_OPENGL)
	if err != nil {
		return nil, err
	}

	r, err := sdl.CreateRenderer(w, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return nil, err
	}

	execPath, err := os.Executable()
	if err != nil {
		return nil, err
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
		return nil, err
	}

	txrUI, err := loadTexture(execPath + "/client/assets/UI.png")
	if err != nil {
		return nil, err
	}

	txrObstacles, err := loadTexture(execPath + "/client/assets/Obstacles.png")
	if err != nil {
		return nil, err
	}

	txrRoads, err := loadTexture(execPath + "/client/assets/Roads.png")
	if err != nil {
		return nil, err
	}

	txrTowers, err := loadTexture(execPath + "/client/assets/Towers.png")
	if err != nil {
		return nil, err
	}

	txrEnemies, err := loadTexture(execPath + "/client/assets/Enemies.png")
	if err != nil {
		return nil, err
	}

	return &SDL{
		game: gm, pid: pid,

		window: w, renderer: r,
		windowW: tileSize * int32(gm.GC.FieldWidth), windowH: tileSize * int32(gm.GC.FieldHeight),
		tileW: tileSize, tileH: tileSize,

		selectedX: 0, selectedY: 0,
		viewOffsetX: 0, viewOffsetY: 0,
		selectedTower: 0,

		textures: textures{
			text:      txrText,
			ui:        txrUI,
			obstacles: txrObstacles,
			roads:     txrRoads,
			towers:    txrTowers,
			enemies:   txrEnemies,
		},
	}, nil
}

func (cl *SDL) Stop() {
	if cl.window != nil {
		_ = cl.window.Destroy()
		cl.window = nil
	}
	if cl.renderer != nil {
		_ = cl.renderer.Destroy()
		cl.renderer = nil
	}
	if cl.textures.obstacles != nil {
		_ = cl.textures.obstacles.Destroy()
		cl.textures.obstacles = nil
	}
	if cl.textures.roads != nil {
		_ = cl.textures.roads.Destroy()
		cl.textures.roads = nil
	}
	if cl.textures.towers != nil {
		_ = cl.textures.towers.Destroy()
		cl.textures.towers = nil
	}
	if cl.textures.enemies != nil {
		_ = cl.textures.enemies.Destroy()
		cl.textures.enemies = nil
	}
}

func (cl *SDL) Draw(processTime time.Duration) error {
	if err := cl.renderer.SetDrawColor(87, 87, 87, 255); err != nil {
		return err
	}
	if err := cl.renderer.Clear(); err != nil {
		return err
	}

	if err := cl.drawField(); err != nil {
		return err
	}
	if err := cl.drawUI(processTime); err != nil {
		return err
	}

	cl.renderer.Present()
	return nil
}

func (cl *SDL) Input() error {
	event := sdl.WaitEventTimeout(100)
	switch event.(type) {
	case *sdl.QuitEvent:
		return game.Errors.Exit

	case *sdl.KeyboardEvent:
		if event.(*sdl.KeyboardEvent).State == 1 {
			return nil
		}

		switch event.(*sdl.KeyboardEvent).Keysym.Scancode {
		case sdl.SCANCODE_ESCAPE:
			return game.Errors.Exit
		case sdl.SCANCODE_P, sdl.SCANCODE_Q:
			cl.game.TogglePause()
			return nil
		case sdl.SCANCODE_RETURN, sdl.SCANCODE_KP_ENTER:
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
		case sdl.SCANCODE_BACKSPACE, sdl.SCANCODE_DELETE:
			return cl.game.StartRound()

		case sdl.SCANCODE_W, sdl.SCANCODE_K:
			cl.selectedY = max(cl.selectedY-1, max(0, -cl.viewOffsetY))
			return nil
		case sdl.SCANCODE_S, sdl.SCANCODE_J:
			cl.selectedY = min(cl.selectedY+1, (cl.game.GC.FieldHeight+min(0, -cl.viewOffsetY))-1)
			return nil
		case sdl.SCANCODE_D, sdl.SCANCODE_L:
			cl.selectedX = min(cl.selectedX+1, (cl.game.GC.FieldWidth+min(0, -cl.viewOffsetX))-1)
			return nil
		case sdl.SCANCODE_A, sdl.SCANCODE_H:
			cl.selectedX = max(cl.selectedX-1, max(0, -cl.viewOffsetX))
			return nil

		case sdl.SCANCODE_UP:
			cl.viewOffsetY = min(cl.viewOffsetY+1, (cl.game.GC.FieldHeight-min(int(cl.windowH/cl.tileH), cl.game.GC.FieldHeight))+6)
			cl.selectedY = max(cl.selectedY-1, max(0, -cl.viewOffsetY))
			return nil
		case sdl.SCANCODE_DOWN:
			cl.viewOffsetY = max(cl.viewOffsetY-1, -5)
			cl.selectedY = min(cl.selectedY+1, (cl.game.GC.FieldHeight+min(0, -cl.viewOffsetY))-1)
			return nil
		case sdl.SCANCODE_RIGHT:
			cl.viewOffsetX = max(cl.viewOffsetX-1, -5)
			cl.selectedX = min(cl.selectedX+1, (cl.game.GC.FieldWidth+min(0, -cl.viewOffsetX))-1)
			return nil
		case sdl.SCANCODE_LEFT:
			cl.viewOffsetX = min(cl.viewOffsetX+1, (cl.game.GC.FieldWidth-min(int(cl.windowW/cl.tileW), cl.game.GC.FieldWidth))+5)
			cl.selectedX = max(cl.selectedX-1, max(0, -cl.viewOffsetX))
			return nil

		case sdl.SCANCODE_LEFTBRACKET, sdl.SCANCODE_MINUS, sdl.SCANCODE_KP_MINUS:
			cl.selectedTower = max(cl.selectedTower-1, 0)
			return nil
		case sdl.SCANCODE_RIGHTBRACKET, sdl.SCANCODE_EQUALS, sdl.SCANCODE_KP_PLUS:
			cl.selectedTower = min(cl.selectedTower+1, len(game.Towers)-1)
			return nil
		}
	}

	return nil
}

func (cl *SDL) renderString(str string, x, y int32) error {
	sources := []sdl.Rect{}
	for _, char := range []rune(str) {
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
			dst := sdl.Rect{X: int32((x + cl.viewOffsetX) * 64), Y: int32((y + cl.viewOffsetY) * 64), W: cl.tileW, H: cl.tileH}

			if err := cl.renderer.FillRect(&dst); err != nil {
				return err
			}

			for _, obj := range cl.game.GetCollisions(x, y) {
				sheet, dstOffset, src := cl.textures.obstacles, dst, sdl.Rect{X: cl.tileW * -1, Y: cl.tileH * -1, W: cl.tileW, H: cl.tileH}

				switch obj.Type() {
				case "Obstacle":
					sheet = cl.textures.obstacles
					switch obj.(*game.ObstacleObj).Name {
					case "lake":
						src = sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
					case "sea":
						src = sdl.Rect{X: cl.tileW * 1, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
					case "sand":
						src = sdl.Rect{X: cl.tileW * 2, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
					case "hills":
						src = sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH}
					case "tree":
						src = sdl.Rect{X: cl.tileW * 1, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH}
					case "brick":
						src = sdl.Rect{X: cl.tileW * 2, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH}
					}

				case "Road":
					sheet = cl.textures.roads
					if obj.(*game.RoadObj).Index == 0 {
						switch obj.(*game.RoadObj).DirExit {
						case "up":
							src = sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH}
						case "right":
							src = sdl.Rect{X: cl.tileW * 1, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH}
						case "down":
							src = sdl.Rect{X: cl.tileW * 2, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH}
						case "left":
							src = sdl.Rect{X: cl.tileW * 3, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH}
						}
					} else if obj.(*game.RoadObj).Index == len(cl.game.GS.Roads)-1 {
						switch obj.(*game.RoadObj).DirEntrance {
						case "up":
							src = sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH}
						case "right":
							src = sdl.Rect{X: cl.tileW * 1, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH}
						case "down":
							src = sdl.Rect{X: cl.tileW * 2, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH}
						case "left":
							src = sdl.Rect{X: cl.tileW * 3, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH}
						}
					} else {
						switch obj.(*game.RoadObj).DirEntrance + ";" + obj.(*game.RoadObj).DirExit {
						case "up;down", "down;up":
							src = sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
						case "left;right", "right;left":
							src = sdl.Rect{X: cl.tileW * 1, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
						case "up;right", "right;up":
							src = sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH}
						case "right;down", "down;right":
							src = sdl.Rect{X: cl.tileW * 1, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH}
						case "down;left", "left;down":
							src = sdl.Rect{X: cl.tileW * 2, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH}
						case "left;up", "up;left":
							src = sdl.Rect{X: cl.tileW * 3, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH}
						}
					}

				case "Tower":
					// TODO: Tower oriantation.

					sheet = cl.textures.towers
					switch obj.(*game.TowerObj).Name {
					case "Basic":
						src = sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
					case "LongRange":
						src = sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 1, W: cl.tileW, H: cl.tileH}
					case "Fast":
						src = sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 2, W: cl.tileW, H: cl.tileH}
					case "Strong":
						src = sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 3, W: cl.tileW, H: cl.tileH}
					}

				case "Enemy":
					if obj.(*game.EnemyObj).Progress < 0.5 {
						continue
					}

					sheet = cl.textures.enemies
					road := cl.game.GS.Roads[min(int(obj.(*game.EnemyObj).Progress), len(cl.game.GS.Roads)-1)]
					switch road.DirEntrance + ";" + road.DirExit {
					case "up;down":
						src = sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
					case "up;left", "right;down":
						src = sdl.Rect{X: cl.tileW * 1, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
					case "right;left":
						src = sdl.Rect{X: cl.tileW * 2, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
					case "right;up", "down;left":
						src = sdl.Rect{X: cl.tileW * 3, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
					case "down;up":
						src = sdl.Rect{X: cl.tileW * 4, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
					case "down;right", "left;up":
						src = sdl.Rect{X: cl.tileW * 5, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
					case "left;right":
						src = sdl.Rect{X: cl.tileW * 6, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
					case "left;down", "up;right":
						src = sdl.Rect{X: cl.tileW * 7, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
					}

					offset := obj.(*game.EnemyObj).Progress - float64(int(obj.(*game.EnemyObj).Progress))

					switch road.DirExit {
					case "up":
						dstOffset.Y -= int32(float64(cl.tileH) * offset)
					case "right":
						dstOffset.X += int32(float64(cl.tileW) * offset)
					case "down":
						dstOffset.Y += int32(float64(cl.tileH) * offset)
					case "left":
						dstOffset.X -= int32(float64(cl.tileW) * offset)
					}
				}

				if err := cl.renderer.Copy(sheet, &src, &dstOffset); err != nil {
					return err
				}
			}

			if x == cl.selectedX && y == cl.selectedY {
				if err := cl.renderer.Copy(cl.textures.ui, &sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}, &dst); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (cl *SDL) drawUI(processTime time.Duration) error {
	phase := "R:" + strconv.Itoa(cl.game.GS.Round+1)
	if cl.game.GS.Phase == "defending" {
		phase += " E:" + strconv.Itoa(len(cl.game.GS.Enemies))
	}

	if err := cl.renderString(phase, 0, 0); err != nil {
		return err
	}

	// if processTime >= cl.game.GC.TickDelay {
	// }
	stats := fmt.Sprintf("%v %v %v", processTime.Milliseconds(), cl.game.Players[cl.pid].Coins, cl.game.GS.Health)
	stats = strings.Repeat(" ", int(cl.windowW/32)-len(stats)-1) + stats

	if err := cl.renderString(stats, 0, 0); err != nil {
		return err
	}

	if cl.game.GS.Phase == "building" {
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
