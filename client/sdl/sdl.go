package clsdl

import (
	"ATowerDefense/game"
	"embed"
	"fmt"
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

		lastMiddleMouseMotion time.Time
	}
)

func newTextures(r *sdl.Renderer, assets embed.FS) (textures, error) {
	loadTexture := func(file string) (*sdl.Texture, error) {
		data, err := assets.ReadFile(file)
		if err != nil {
			return nil, err
		}
		rw, err := sdl.RWFromMem(data)
		if err != nil {
			return nil, err
		}
		defer func() { _ = rw.Free() }()
		srf, err := img.LoadPNGRW(rw)
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

	txrText, err := loadTexture("assets/Text.png")
	if err != nil {
		return textures{}, err
	}
	txrUI, err := loadTexture("assets/UI.png")
	if err != nil {
		return textures{}, err
	}
	txrObstacles, err := loadTexture("assets/Obstacles.png")
	if err != nil {
		return textures{}, err
	}
	txrRoads, err := loadTexture("assets/Roads.png")
	if err != nil {
		return textures{}, err
	}
	txrTowers, err := loadTexture("assets/Towers.png")
	if err != nil {
		return textures{}, err
	}
	txrEnemies, err := loadTexture("assets/Enemies.png")
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

func NewSDL(gm *game.Game, pid int, assets embed.FS) (*SDL, error) {
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
	if err := r.SetDrawBlendMode(sdl.BLENDMODE_BLEND); err != nil {
		return nil, err
	}

	textures, err := newTextures(r, assets)
	if err != nil {
		return nil, err
	}

	return &SDL{
		game: gm, pid: pid,

		window: w, renderer: r,
		windowW: tileSize * int32(gm.GC.FieldWidth), windowH: tileSize * int32(gm.GC.FieldHeight),
		tileW: tileSize, tileH: tileSize,

		selectedX: gm.GC.FieldWidth / 2, selectedY: gm.GC.FieldHeight / 2,
		viewOffsetX: 0, viewOffsetY: 0,
		selectedTower: 0,

		textures: textures,

		lastMiddleMouseMotion: time.Now(),
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
	switch event := event.(type) {
	case *sdl.QuitEvent:
		return game.Errors.Exit

	case *sdl.KeyboardEvent:
		if event.State != sdl.PRESSED {
			return nil
		}

		switch event.Keysym.Scancode {
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

	case *sdl.MouseMotionEvent:
		switch event.State {
		case sdl.BUTTON_MIDDLE:
			if time.Since(cl.lastMiddleMouseMotion) < time.Millisecond*25 {
				return nil
			}
			cl.lastMiddleMouseMotion = time.Now()

			fmt.Println(event.XRel, event.YRel)
			if event.XRel > 1 {
				cl.viewOffsetX = min(cl.viewOffsetX+1, (cl.game.GC.FieldWidth-min(int(cl.windowW/cl.tileW), cl.game.GC.FieldWidth))+5)
			} else if event.XRel < -1 {
				cl.viewOffsetX = max(cl.viewOffsetX-1, -5)
			}

			if event.YRel > 1 {
				cl.viewOffsetY = min(cl.viewOffsetY+1, (cl.game.GC.FieldHeight-min(int(cl.windowH/cl.tileH), cl.game.GC.FieldHeight))+6)
			} else if event.YRel < -1 {
				cl.viewOffsetY = max(cl.viewOffsetY-1, -5)
			}

			return nil

		default:
			cl.selectedX = min(max(int(event.X/cl.tileW)-cl.viewOffsetX, 0), (cl.game.GC.FieldWidth+min(0, -cl.viewOffsetX))-1)
			cl.selectedY = min(max(int(event.Y/cl.tileH)-cl.viewOffsetY, 0), (cl.game.GC.FieldHeight+min(0, -cl.viewOffsetY))-1)
			return nil
		}

	case *sdl.MouseButtonEvent:
		if event.State != sdl.RELEASED {
			return nil
		}

		switch event.Button {
		case sdl.BUTTON_LEFT:
			if len(game.Towers) < cl.selectedTower {
				return nil
			}
			return cl.game.PlaceTower(game.Towers[cl.selectedTower].Name, cl.selectedX, cl.selectedY, cl.pid)
		case sdl.BUTTON_RIGHT:
			return cl.game.DestoryTower(cl.selectedX, cl.selectedY, cl.pid)

		case sdl.BUTTON_X1, sdl.BUTTON_X2:
			if cl.game.GS.Phase == "defending" {
				cl.game.TogglePause()
			} else {
				return cl.game.StartRound()
			}
			return nil
		}

	case *sdl.MouseWheelEvent:
		fmt.Println(event.X, event.Y)
		if event.Y > 0 {
			cl.selectedTower = max(cl.selectedTower-1, 0)
			return nil
		} else if event.Y < 0 {
			cl.selectedTower = min(cl.selectedTower+1, len(game.Towers)-1)
			return nil
		}
	}

	return nil
}
