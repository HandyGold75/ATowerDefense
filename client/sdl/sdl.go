package clsdl

import (
	"ATowerDefense/game"
	"fmt"
	"strconv"
	"strings"
	"time"

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
	if err := r.SetDrawBlendMode(sdl.BLENDMODE_BLEND); err != nil {
		return nil, err
	}

	textures, err := newTextures(r)
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

		textures: textures,
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
		if event.State != 1 {
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
				sheet, dstOffset, src := cl.textures.obstacles, dst, cl.newRect(0, 0, 0, 0)
				switch obj := obj.(type) {
				case *game.ObstacleObj:
					sheet, src = cl.textures.obstacles, cl.srcObstacle(obj)

				case *game.RoadObj:
					sheet, src = cl.textures.roads, cl.srcRoad(obj)

				case *game.TowerObj:
					sheet, src = cl.textures.towers, cl.srcTower(obj)

				case *game.EnemyObj:
					if obj.Progress < 0.5 {
						continue
					}
					sheet, src = cl.textures.enemies, cl.srcEnemy(obj, &dstOffset)
				}

				if err := cl.renderer.Copy(sheet, &src, &dstOffset); err != nil {
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
