package clsdl

import (
	"ATowerDefense/game"
	"os"
	"strings"
	"time"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
)

type (
	textures struct {
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
				sheet, src := cl.textures.obstacles, sdl.Rect{X: cl.tileW * -1, Y: cl.tileH * -1, W: cl.tileW, H: cl.tileH}

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
					road := cl.game.GS.Roads[max(int(obj.(*game.EnemyObj).Progress), len(cl.game.GS.Roads)-1)]
					switch road.DirEntrance + ";" + road.DirExit {
					case "up;down": //
						src = sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
					case "up;left", "right;down":
						src = sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
					case "right;left": //
						src = sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
					case "right;up", "down;left":
						src = sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
					case "down;up": //
						src = sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
					case "down;right", "left;up":
						src = sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
					case "left;right": //
						src = sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
					case "left;down", "up;right":
						src = sdl.Rect{X: cl.tileW * 0, Y: cl.tileH * 0, W: cl.tileW, H: cl.tileH}
					}
				}

				if err := cl.renderer.Copy(sheet, &src, &dst); err != nil {
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
	return nil
}

func (cl *SDL) Draw(processTime time.Duration) error {
	if err := cl.renderer.SetDrawColor(0, 0, 0, 255); err != nil {
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
