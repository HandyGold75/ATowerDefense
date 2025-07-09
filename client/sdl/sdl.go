package clsdl

import (
	"ATowerDefense/game"
	"embed"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

type (
	textures struct {
		text        *sdl.Texture
		ui          *sdl.Texture
		environment *sdl.Texture
		roads       *sdl.Texture
		towers      *sdl.Texture
		enemies     *sdl.Texture
	}

	SDL struct {
		GM  *game.Game
		pid int

		window   *sdl.Window
		renderer *sdl.Renderer
		assets   embed.FS

		windowW, windowH,
		tileW, tileH int32

		selectedX, selectedY,
		viewOffsetX, viewOffsetY,
		selectedTower int

		theme    string
		themeNew string
		Textures textures

		warningMsg            string
		warningMsgTimeout     time.Time
		lastMiddleMouseMotion time.Time
	}
)

func NewSDL(gc game.GameConfig, assets embed.FS) (*SDL, error) {
	gm := game.NewGame(gc)
	if err := gm.Start(); err != nil {
		return nil, err
	}
	pid := gm.AddPlayer()

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return nil, err
	}

	tileSize, theme := int32(64), "city"
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

	cl := &SDL{
		GM: gm, pid: pid,

		window: w, renderer: r, assets: assets,
		windowW: tileSize * int32(gm.GC.FieldWidth), windowH: tileSize * int32(gm.GC.FieldHeight),
		tileW: tileSize, tileH: tileSize,

		selectedX: gm.GC.FieldWidth / 2, selectedY: gm.GC.FieldHeight / 2,
		viewOffsetX: 0, viewOffsetY: 0,
		selectedTower: 0,

		theme: theme, themeNew: theme, Textures: textures{},

		lastMiddleMouseMotion: time.Now(),
	}
	if err := cl.loadTheme("old"); err != nil {
		return nil, err
	}

	return cl, nil
}

func (cl *SDL) Stop() {
	if cl.GM.GS.State != "stopped" {
		_ = cl.GM.Stop()
	}

	if cl.window != nil {
		_ = cl.window.Destroy()
		cl.window = nil
	}
	if cl.renderer != nil {
		_ = cl.renderer.Destroy()
		cl.renderer = nil
	}

	if cl.Textures.text != nil {
		_ = cl.Textures.text.Destroy()
		cl.Textures.text = nil
	}
	if cl.Textures.ui != nil {
		_ = cl.Textures.ui.Destroy()
		cl.Textures.ui = nil
	}
	if cl.Textures.environment != nil {
		_ = cl.Textures.environment.Destroy()
		cl.Textures.environment = nil
	}
	if cl.Textures.roads != nil {
		_ = cl.Textures.roads.Destroy()
		cl.Textures.roads = nil
	}
	if cl.Textures.towers != nil {
		_ = cl.Textures.towers.Destroy()
		cl.Textures.towers = nil
	}
	if cl.Textures.enemies != nil {
		_ = cl.Textures.enemies.Destroy()
		cl.Textures.enemies = nil
	}
}

func (cl *SDL) Draw(processTime time.Duration) error {
	if cl.theme != cl.themeNew {
		if err := cl.loadTheme(cl.themeNew); err != nil {
			return err
		}
	}

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

	if time.Until(cl.warningMsgTimeout) > 0 {
		if err := cl.renderString(cl.warningMsg, (cl.windowW/2)-(cl.tileW/2)-((cl.tileW/2)*int32(len(cl.warningMsg)/2)), (cl.windowH)-(cl.tileH)); err != nil {
			return err
		}
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
			cl.GM.TogglePause()
		case sdl.SCANCODE_BACKSPACE, sdl.SCANCODE_DELETE:
			if err := cl.GM.StartRound(); err != nil {
				cl.warningMsg = err.Error()
				cl.warningMsgTimeout = time.Now().Add(time.Second * 3)
			}

		case sdl.SCANCODE_RETURN, sdl.SCANCODE_KP_ENTER:
			if len(game.Towers) < cl.selectedTower {
				return nil
			}
			if err := cl.GM.PlaceTower(game.Towers[cl.selectedTower].Name, cl.selectedX, cl.selectedY, cl.pid); err != nil {
				if err != game.Errors.InvalidPlacement {
					cl.warningMsg = err.Error()
					cl.warningMsgTimeout = time.Now().Add(time.Second * 3)
				} else if err := cl.GM.DestoryObstacle(cl.selectedX, cl.selectedY, cl.pid); err != nil {
					if err != game.Errors.InvalidPlacement {
						cl.warningMsg = err.Error()
						cl.warningMsgTimeout = time.Now().Add(time.Second * 3)
					} else if err := cl.GM.DestoryTower(cl.selectedX, cl.selectedY, cl.pid); err != nil {
						cl.warningMsg = err.Error()
						cl.warningMsgTimeout = time.Now().Add(time.Second * 3)
					}
				}
			}
		case sdl.SCANCODE_T:
			switch cl.theme {
			case "old":
				cl.themeNew = "city"
			case "city":
				cl.themeNew = "old"
			}

		case sdl.SCANCODE_W, sdl.SCANCODE_K:
			cl.selectedY = max(cl.selectedY-1, max(0, -cl.viewOffsetY))
		case sdl.SCANCODE_S, sdl.SCANCODE_J:
			cl.selectedY = min(cl.selectedY+1, (cl.GM.GC.FieldHeight+min(0, -cl.viewOffsetY))-1)
		case sdl.SCANCODE_D, sdl.SCANCODE_L:
			cl.selectedX = min(cl.selectedX+1, (cl.GM.GC.FieldWidth+min(0, -cl.viewOffsetX))-1)
		case sdl.SCANCODE_A, sdl.SCANCODE_H:
			cl.selectedX = max(cl.selectedX-1, max(0, -cl.viewOffsetX))

		case sdl.SCANCODE_UP:
			cl.viewOffsetY = min(cl.viewOffsetY+1, (cl.GM.GC.FieldHeight-min(int(cl.windowH/cl.tileH), cl.GM.GC.FieldHeight))+6)
			cl.selectedY = max(cl.selectedY-1, max(0, -cl.viewOffsetY))
		case sdl.SCANCODE_DOWN:
			cl.viewOffsetY = max(cl.viewOffsetY-1, -5)
			cl.selectedY = min(cl.selectedY+1, (cl.GM.GC.FieldHeight+min(0, -cl.viewOffsetY))-1)
		case sdl.SCANCODE_RIGHT:
			cl.viewOffsetX = max(cl.viewOffsetX-1, -5)
			cl.selectedX = min(cl.selectedX+1, (cl.GM.GC.FieldWidth+min(0, -cl.viewOffsetX))-1)
		case sdl.SCANCODE_LEFT:
			cl.viewOffsetX = min(cl.viewOffsetX+1, (cl.GM.GC.FieldWidth-min(int(cl.windowW/cl.tileW), cl.GM.GC.FieldWidth))+5)
			cl.selectedX = max(cl.selectedX-1, max(0, -cl.viewOffsetX))

		case sdl.SCANCODE_LEFTBRACKET:
			cl.selectedTower = max(cl.selectedTower-1, 0)
		case sdl.SCANCODE_RIGHTBRACKET:
			cl.selectedTower = min(cl.selectedTower+1, len(game.Towers)-1)

		case sdl.SCANCODE_EQUALS, sdl.SCANCODE_KP_PLUS:
			cl.GM.GC.GameSpeed = min(cl.GM.GC.GameSpeed+1, 9)
		case sdl.SCANCODE_MINUS, sdl.SCANCODE_KP_MINUS:
			cl.GM.GC.GameSpeed = max(cl.GM.GC.GameSpeed-1, 0)
		}

		return nil

	case *sdl.MouseMotionEvent:
		switch event.State {
		case sdl.BUTTON_MIDDLE:
			if time.Since(cl.lastMiddleMouseMotion) < time.Millisecond*50 {
				return nil
			}
			cl.lastMiddleMouseMotion = time.Now()

			if event.XRel > 0 {
				cl.viewOffsetX = min(cl.viewOffsetX+1, (cl.GM.GC.FieldWidth-min(int(cl.windowW/cl.tileW), cl.GM.GC.FieldWidth))+5)
			} else if event.XRel < 0 {
				cl.viewOffsetX = max(cl.viewOffsetX-1, -5)
			}

			if event.YRel > 0 {
				cl.viewOffsetY = min(cl.viewOffsetY+1, (cl.GM.GC.FieldHeight-min(int(cl.windowH/cl.tileH), cl.GM.GC.FieldHeight))+6)
			} else if event.YRel < 0 {
				cl.viewOffsetY = max(cl.viewOffsetY-1, -5)
			}

		default:
			cl.selectedX = min(max(int(event.X/cl.tileW)-cl.viewOffsetX, 0), (cl.GM.GC.FieldWidth+min(0, -cl.viewOffsetX))-1)
			cl.selectedY = min(max(int(event.Y/cl.tileH)-cl.viewOffsetY, 0), (cl.GM.GC.FieldHeight+min(0, -cl.viewOffsetY))-1)
		}

		return nil

	case *sdl.MouseButtonEvent:
		if event.State != sdl.RELEASED {
			return nil
		}

		switch event.Button {
		case sdl.BUTTON_LEFT:
			if len(game.Towers) < cl.selectedTower {
				return nil
			}
			if err := cl.GM.PlaceTower(game.Towers[cl.selectedTower].Name, cl.selectedX, cl.selectedY, cl.pid); err != nil {
				cl.warningMsg = err.Error()
				cl.warningMsgTimeout = time.Now().Add(time.Second * 3)
			}
		case sdl.BUTTON_RIGHT:
			if err := cl.GM.DestoryObstacle(cl.selectedX, cl.selectedY, cl.pid); err != nil {
				if err != game.Errors.InvalidPlacement {
					cl.warningMsg = err.Error()
					cl.warningMsgTimeout = time.Now().Add(time.Second * 3)
				} else if err := cl.GM.DestoryTower(cl.selectedX, cl.selectedY, cl.pid); err != nil {
					cl.warningMsg = err.Error()
					cl.warningMsgTimeout = time.Now().Add(time.Second * 3)
				}
			}

		case sdl.BUTTON_X1, sdl.BUTTON_X2:
			if cl.GM.GS.Phase == "defending" {
				cl.GM.TogglePause()
			} else {
				if err := cl.GM.StartRound(); err != nil {
					cl.warningMsg = err.Error()
					cl.warningMsgTimeout = time.Now().Add(time.Second * 3)
				}
			}
		}

		return nil

	case *sdl.MouseWheelEvent:
		if event.Y > 0 {
			cl.selectedTower = max(cl.selectedTower-1, 0)
		} else if event.Y < 0 {
			cl.selectedTower = min(cl.selectedTower+1, len(game.Towers)-1)
		}
		return nil
	}

	return nil
}

func (cl *SDL) loadTheme(theme string) error {
	loadTexture := func(file string) (*sdl.Texture, error) {
		var rw *sdl.RWops
		if data, err := cl.assets.ReadFile("assets/" + file); err == nil {
			rw, _ = sdl.RWFromMem(data)
		}
		if rw == nil {
			rw = sdl.RWFromFile("client/assets/"+file, "rb")
		}
		defer func() { _ = rw.Free() }()

		srf, err := img.LoadPNGRW(rw)
		if err != nil {
			return nil, err
		}
		defer srf.Free()

		txr, err := cl.renderer.CreateTextureFromSurface(srf)
		if err != nil {
			return nil, err
		}

		return txr, nil
	}

	txrText, err := loadTexture(theme + "/Text.png")
	if err != nil {
		return err
	}
	txrUI, err := loadTexture(theme + "/UI.png")
	if err != nil {
		return err
	}
	txrEnvironment, err := loadTexture(theme + "/Environment.png")
	if err != nil {
		return err
	}
	txrRoads, err := loadTexture(theme + "/Roads.png")
	if err != nil {
		return err
	}
	txrTowers, err := loadTexture(theme + "/Towers.png")
	if err != nil {
		return err
	}
	txrEnemies, err := loadTexture(theme + "/Enemies.png")
	if err != nil {
		return err
	}

	if cl.Textures.text != nil {
		if err := cl.Textures.text.Destroy(); err != nil {
			return err
		}
		cl.Textures.text = nil
	}
	if cl.Textures.ui != nil {
		if err := cl.Textures.ui.Destroy(); err != nil {
			return err
		}
		cl.Textures.ui = nil
	}
	if cl.Textures.environment != nil {
		if err := cl.Textures.environment.Destroy(); err != nil {
			return err
		}
		cl.Textures.environment = nil
	}
	if cl.Textures.roads != nil {
		if err := cl.Textures.roads.Destroy(); err != nil {
			return err
		}
		cl.Textures.roads = nil
	}
	if cl.Textures.towers != nil {
		if err := cl.Textures.towers.Destroy(); err != nil {
			return err
		}
		cl.Textures.towers = nil
	}
	if cl.Textures.enemies != nil {
		if err := cl.Textures.enemies.Destroy(); err != nil {
			return err
		}
		cl.Textures.enemies = nil
	}

	cl.theme = theme
	cl.Textures = textures{
		text:        txrText,
		ui:          txrUI,
		environment: txrEnvironment,
		roads:       txrRoads,
		towers:      txrTowers,
		enemies:     txrEnemies,
	}

	return nil
}
