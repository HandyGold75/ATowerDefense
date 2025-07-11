package clsdl

import (
	"ATowerDefense/game"
	"embed"
	"fmt"
	"math"
	"math/rand/v2"
	"strconv"
	"strings"
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

	clSDL struct {
		gm  *game.Game
		pid int

		window   *sdl.Window
		renderer *sdl.Renderer
		assets   embed.FS

		windowW, windowH int32

		selectedX, selectedY,
		viewOffsetX, viewOffsetY,
		selectedTower int

		theme    string
		themeNew string
		textures textures

		warningMsg            string
		warningMsgTimeout     time.Time
		lastMiddleMouseMotion time.Time
	}
)

var (
	backgroundCache = map[int]map[int]sdl.Rect{}
	obstacleCache   = map[int]sdl.Rect{}

	// 0.0 - 0.5; lower makes the rotate anamation longer
	rotateAnimationOffset = float64(1) / 3
)

func Run(gc game.GameConfig, assets embed.FS) error {
	cl, err := newSDL(gc, assets)
	if err != nil {
		return err
	}
	defer cl.stop()
	cl.start()
	return err
}

func newSDL(gc game.GameConfig, assets embed.FS) (*clSDL, error) {
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

	cl := &clSDL{
		gm: gm, pid: pid,

		window: w, renderer: r, assets: assets,
		windowW: tileSize * int32(gm.GC.FieldWidth), windowH: tileSize * int32(gm.GC.FieldHeight),

		selectedX: gm.GC.FieldWidth / 2, selectedY: gm.GC.FieldHeight / 2,
		viewOffsetX: 0, viewOffsetY: 0,
		selectedTower: 0,

		theme: theme, themeNew: theme, textures: textures{},

		lastMiddleMouseMotion: time.Now(),
	}
	if err := cl.loadTheme(theme); err != nil {
		return nil, err
	}

	return cl, nil
}

func (cl *clSDL) start() {
	go func() {
		defer cl.stop()
		for cl.gm.GS.State != "stopped" {
			if err := cl.input(); err != nil {
				if err == game.Errors.Exit {
					break
				}
				fmt.Println(err)
			}
		}
	}()
	if err := cl.gm.Run(cl.draw); err != nil {
		fmt.Println(err)
	}
}

func (cl *clSDL) stop() {
	if cl.gm.GS.State != "stopped" {
		_ = cl.gm.Stop()
	}

	if cl.window != nil {
		_ = cl.window.Destroy()
		cl.window = nil
	}
	if cl.renderer != nil {
		_ = cl.renderer.Destroy()
		cl.renderer = nil
	}

	if cl.textures.text != nil {
		_ = cl.textures.text.Destroy()
		cl.textures.text = nil
	}
	if cl.textures.ui != nil {
		_ = cl.textures.ui.Destroy()
		cl.textures.ui = nil
	}
	if cl.textures.environment != nil {
		_ = cl.textures.environment.Destroy()
		cl.textures.environment = nil
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

func (cl *clSDL) draw(processTime time.Duration) error {
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
		if err := cl.renderString(cl.warningMsg, (cl.windowW/2)-(tileSize/2)-((tileSize/2)*int32(len(cl.warningMsg)/2)), (cl.windowH)-(tileSize)); err != nil {
			return err
		}
	}

	cl.renderer.Present()
	return nil
}

func (cl *clSDL) input() error {
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
			cl.gm.TogglePause()
		case sdl.SCANCODE_BACKSPACE, sdl.SCANCODE_DELETE:
			if err := cl.gm.StartRound(); err != nil {
				cl.warningMsg = err.Error()
				cl.warningMsgTimeout = time.Now().Add(time.Second * 3)
			}

		case sdl.SCANCODE_RETURN, sdl.SCANCODE_KP_ENTER:
			if len(game.Towers) < cl.selectedTower {
				return nil
			}
			if err := cl.gm.PlaceTower(game.Towers[cl.selectedTower].Name, cl.selectedX, cl.selectedY, cl.pid); err != nil {
				if err != game.Errors.InvalidPlacement {
					cl.warningMsg = err.Error()
					cl.warningMsgTimeout = time.Now().Add(time.Second * 3)
				} else if err := cl.gm.DestroyObstacle(cl.selectedX, cl.selectedY, cl.pid); err != nil {
					if err != game.Errors.InvalidPlacement {
						cl.warningMsg = err.Error()
						cl.warningMsgTimeout = time.Now().Add(time.Second * 3)
					} else if err := cl.gm.DestroyTower(cl.selectedX, cl.selectedY, cl.pid); err != nil {
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
			cl.selectedY = min(cl.selectedY+1, (cl.gm.GC.FieldHeight+min(0, -cl.viewOffsetY))-1)
		case sdl.SCANCODE_D, sdl.SCANCODE_L:
			cl.selectedX = min(cl.selectedX+1, (cl.gm.GC.FieldWidth+min(0, -cl.viewOffsetX))-1)
		case sdl.SCANCODE_A, sdl.SCANCODE_H:
			cl.selectedX = max(cl.selectedX-1, max(0, -cl.viewOffsetX))

		case sdl.SCANCODE_UP:
			cl.viewOffsetY = min(cl.viewOffsetY+1, (cl.gm.GC.FieldHeight-min(int(cl.windowH/tileSize), cl.gm.GC.FieldHeight))+6)
			cl.selectedY = max(cl.selectedY-1, max(0, -cl.viewOffsetY))
		case sdl.SCANCODE_DOWN:
			cl.viewOffsetY = max(cl.viewOffsetY-1, -5)
			cl.selectedY = min(cl.selectedY+1, (cl.gm.GC.FieldHeight+min(0, -cl.viewOffsetY))-1)
		case sdl.SCANCODE_RIGHT:
			cl.viewOffsetX = max(cl.viewOffsetX-1, -5)
			cl.selectedX = min(cl.selectedX+1, (cl.gm.GC.FieldWidth+min(0, -cl.viewOffsetX))-1)
		case sdl.SCANCODE_LEFT:
			cl.viewOffsetX = min(cl.viewOffsetX+1, (cl.gm.GC.FieldWidth-min(int(cl.windowW/tileSize), cl.gm.GC.FieldWidth))+5)
			cl.selectedX = max(cl.selectedX-1, max(0, -cl.viewOffsetX))

		case sdl.SCANCODE_LEFTBRACKET:
			cl.selectedTower = max(cl.selectedTower-1, 0)
		case sdl.SCANCODE_RIGHTBRACKET:
			cl.selectedTower = min(cl.selectedTower+1, len(game.Towers)-1)

		case sdl.SCANCODE_EQUALS, sdl.SCANCODE_KP_PLUS:
			cl.gm.GC.GameSpeed = min(cl.gm.GC.GameSpeed+1, 9)
		case sdl.SCANCODE_MINUS, sdl.SCANCODE_KP_MINUS:
			cl.gm.GC.GameSpeed = max(cl.gm.GC.GameSpeed-1, 0)
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
				cl.viewOffsetX = min(cl.viewOffsetX+1, (cl.gm.GC.FieldWidth-min(int(cl.windowW/tileSize), cl.gm.GC.FieldWidth))+5)
			} else if event.XRel < 0 {
				cl.viewOffsetX = max(cl.viewOffsetX-1, -5)
			}

			if event.YRel > 0 {
				cl.viewOffsetY = min(cl.viewOffsetY+1, (cl.gm.GC.FieldHeight-min(int(cl.windowH/tileSize), cl.gm.GC.FieldHeight))+6)
			} else if event.YRel < 0 {
				cl.viewOffsetY = max(cl.viewOffsetY-1, -5)
			}

		default:
			cl.selectedX = min(max(int(event.X/tileSize)-cl.viewOffsetX, 0), (cl.gm.GC.FieldWidth+min(0, -cl.viewOffsetX))-1)
			cl.selectedY = min(max(int(event.Y/tileSize)-cl.viewOffsetY, 0), (cl.gm.GC.FieldHeight+min(0, -cl.viewOffsetY))-1)
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
			if err := cl.gm.PlaceTower(game.Towers[cl.selectedTower].Name, cl.selectedX, cl.selectedY, cl.pid); err != nil {
				cl.warningMsg = err.Error()
				cl.warningMsgTimeout = time.Now().Add(time.Second * 3)
			}
		case sdl.BUTTON_RIGHT:
			if err := cl.gm.DestroyObstacle(cl.selectedX, cl.selectedY, cl.pid); err != nil {
				if err != game.Errors.InvalidPlacement {
					cl.warningMsg = err.Error()
					cl.warningMsgTimeout = time.Now().Add(time.Second * 3)
				} else if err := cl.gm.DestroyTower(cl.selectedX, cl.selectedY, cl.pid); err != nil {
					cl.warningMsg = err.Error()
					cl.warningMsgTimeout = time.Now().Add(time.Second * 3)
				}
			}

		case sdl.BUTTON_X1, sdl.BUTTON_X2:
			if cl.gm.GS.Phase == "defending" {
				cl.gm.TogglePause()
			} else {
				if err := cl.gm.StartRound(); err != nil {
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

func (cl *clSDL) loadTheme(theme string) error {
	loadTexture := func(file string) (*sdl.Texture, error) {
		var rw *sdl.RWops
		if data, err := cl.assets.ReadFile("assets/" + file); err == nil {
			rw, _ = sdl.RWFromMem(data)
		}
		if rw == nil {
			rw = sdl.RWFromFile("assets/"+file, "rb")
			fmt.Println("Warning, theme \"" + file + "\" loaded from disk")
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

	if cl.textures.text != nil {
		if err := cl.textures.text.Destroy(); err != nil {
			return err
		}
		cl.textures.text = nil
	}
	if cl.textures.ui != nil {
		if err := cl.textures.ui.Destroy(); err != nil {
			return err
		}
		cl.textures.ui = nil
	}
	if cl.textures.environment != nil {
		if err := cl.textures.environment.Destroy(); err != nil {
			return err
		}
		cl.textures.environment = nil
	}
	if cl.textures.roads != nil {
		if err := cl.textures.roads.Destroy(); err != nil {
			return err
		}
		cl.textures.roads = nil
	}
	if cl.textures.towers != nil {
		if err := cl.textures.towers.Destroy(); err != nil {
			return err
		}
		cl.textures.towers = nil
	}
	if cl.textures.enemies != nil {
		if err := cl.textures.enemies.Destroy(); err != nil {
			return err
		}
		cl.textures.enemies = nil
	}

	cl.theme = theme
	cl.textures = textures{
		text:        txrText,
		ui:          txrUI,
		environment: txrEnvironment,
		roads:       txrRoads,
		towers:      txrTowers,
		enemies:     txrEnemies,
	}

	return nil
}

func (cl *clSDL) newRect(x, y int32) sdl.Rect {
	return sdl.Rect{X: x * tileSize, Y: y * tileSize, W: tileSize, H: tileSize}
}

func (cl *clSDL) renderString(str string, x, y int32) error {
	sources := []sdl.Rect{}
	for _, char := range str {
		sources = append(sources, textureText[char])
	}

	for i, src := range sources {
		if err := cl.renderer.Copy(cl.textures.text, &src, &sdl.Rect{X: x + ((tileSize / 2) * int32(i)), Y: y, W: tileSize, H: tileSize}); err != nil {
			return err
		}
	}
	return nil
}

func (cl *clSDL) drawField() error {
	for y := range cl.gm.GC.FieldHeight {
		for x := range cl.gm.GC.FieldWidth {
			dst := cl.newRect(int32(x+cl.viewOffsetX), int32(y+cl.viewOffsetY))
			src, ok := backgroundCache[x][y]
			if !ok {
				src = textureBackground[rand.Int32N(6)]
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
		dst := cl.newRect(int32(x+cl.viewOffsetX), int32(y+cl.viewOffsetY))
		src := textureRoads[road.DirEntrance+";"+road.DirExit]
		if err := cl.renderer.Copy(cl.textures.roads, &src, &dst); err != nil {
			return err
		}
	}

	for _, tower := range cl.gm.GS.Towers {
		x, y := tower.Cord()
		dst := cl.newRect(int32(x+cl.viewOffsetX), int32(y+cl.viewOffsetY))
		src := textureTowers[tower.Name][min(int32((tower.Rotation/360)*16), 15)]
		if err := cl.renderer.Copy(cl.textures.towers, &src, &dst); err != nil {
			return err
		}
		dst.Y -= int32(float64(tileSize) * 0.75)
		src = textureUI["barblue;"+strconv.Itoa(int(math.Round(min(tower.ReloadProgress, 1)*9)))]

		if err := cl.renderer.Copy(cl.textures.ui, &src, &dst); err != nil {
			return err
		}
	}

	for _, obstacle := range cl.gm.GS.Obstacles {
		x, y := obstacle.Cord()
		dst := cl.newRect(int32(x+cl.viewOffsetX), int32(y+cl.viewOffsetY))
		src, ok := obstacleCache[obstacle.UID]
		if !ok {
			src = textureObstacles[rand.Int32N(6)]
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
		dst := cl.newRect(int32(x+cl.viewOffsetX), int32(y+cl.viewOffsetY))
		road := cl.gm.GS.Roads[min(int(enemy.Progress), len(cl.gm.GS.Roads)-1)]
		src := textureEnemies[road.DirEntrance+";"+road.DirExit]

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
				dst.Y -= int32(float64(tileSize) * (0.5 - progdec))
			case "right":
				dst.X += int32(float64(tileSize) * (0.5 - progdec))
			case "down":
				dst.Y += int32(float64(tileSize) * (0.5 - progdec))
			case "left":
				dst.X -= int32(float64(tileSize) * (0.5 - progdec))
			}
			src = textureEnemies[road.DirEntrance+";end"]

		case progdec >= 1-rotateAnimationOffset:
			switch road.DirExit {
			case "up":
				dst.Y -= int32(float64(tileSize) * (progdec - 0.5))
			case "right":
				dst.X += int32(float64(tileSize) * (progdec - 0.5))
			case "down":
				dst.Y += int32(float64(tileSize) * (progdec - 0.5))
			case "left":
				dst.X -= int32(float64(tileSize) * (progdec - 0.5))
			}
			src = textureEnemies["start;"+road.DirExit]

		default:
			switch road.DirEntrance {
			case "up":
				dst.Y -= int32(float64(tileSize) * (0.25 - (progdec / 2)))
			case "right":
				dst.X += int32(float64(tileSize) * (0.25 - (progdec / 2)))
			case "down":
				dst.Y += int32(float64(tileSize) * (0.25 - (progdec / 2)))
			case "left":
				dst.X -= int32(float64(tileSize) * (0.25 - (progdec / 2)))
			}
			switch road.DirExit {
			case "up":
				dst.Y -= int32(float64(tileSize) * ((progdec / 2) - 0.25))
			case "right":
				dst.X += int32(float64(tileSize) * ((progdec / 2) - 0.25))
			case "down":
				dst.Y += int32(float64(tileSize) * ((progdec / 2) - 0.25))
			case "left":
				dst.X -= int32(float64(tileSize) * ((progdec / 2) - 0.25))
			}
		}

		if err := cl.renderer.Copy(cl.textures.enemies, &src, &dst); err != nil {
			return err
		}
		dst.Y -= int32(float64(tileSize) * 0.75)
		src = textureUI["barred;"+strconv.Itoa(int(math.Round(float64(enemy.Health)/float64(enemy.StartHealth)*9)))]
		if err := cl.renderer.Copy(cl.textures.ui, &src, &dst); err != nil {
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
		dst := cl.newRect(int32(cl.selectedX+cl.viewOffsetX-r), int32(cl.selectedY+cl.viewOffsetY-r))
		dst.W, dst.H = int32((r*2)+1)*tileSize, int32((r*2)+1)*tileSize
		if err := cl.renderer.FillRect(&dst); err != nil {
			return err
		}
	}

	dst := cl.newRect(int32(cl.selectedX+cl.viewOffsetX), int32(cl.selectedY+cl.viewOffsetY))
	src := textureUI["crosshair"]
	if err := cl.renderer.Copy(cl.textures.ui, &src, &dst); err != nil {
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
			if err := cl.renderString(tower.Name+" <", 0, (cl.windowH-(tileSize*int32(len(game.Towers))))+(tileSize*int32(i))); err != nil {
				return err
			}
			continue
		}
		if err := cl.renderString(tower.Name, 0, (cl.windowH-(tileSize*int32(len(game.Towers))))+(tileSize*int32(i))); err != nil {
			return err
		}
	}

	if cl.gm.GS.State == "paused" {
		msg := "Paused"
		if err := cl.renderString(msg, (cl.windowW/2)-(tileSize/2)-((tileSize/2)*int32(len(msg)/2)), (cl.windowH/2)-(tileSize/2)); err != nil {
			return err
		}
	}

	if cl.gm.GS.Phase == "lost" {
		msg := "Game Over"
		if err := cl.renderString(msg, (cl.windowW/2)-(tileSize/2)-((tileSize/2)*int32(len(msg)/2)), (cl.windowH/2)-(tileSize/2)); err != nil {
			return err
		}
	}

	return nil
}
