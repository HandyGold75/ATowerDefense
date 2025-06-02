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

		windowWidth, windowHeight int32

		selectedX, selectedY,
		selectedTower int

		textures textures
	}
)

func NewSDL(gm *game.Game, pid int) (*SDL, error) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return nil, err
	}

	var width, height int32 = 1920, 1080
	w, err := sdl.CreateWindow("ATowerDefense", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, width, height, sdl.WINDOW_OPENGL)
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

	loadTexture := func(file string) (*sdl.Texture, *sdl.Surface, error) {
		srf, err := img.LoadPNGRW(sdl.RWFromFile(file, "rb"))
		if err != nil {
			return nil, nil, err
		}
		txr, err := r.CreateTextureFromSurface(srf)
		if err != nil {
			return nil, nil, err
		}
		return txr, srf, nil
	}

	txrObstacles, srfObstacles, err := loadTexture(execPath + "/client/assets/Obstacles.png")
	defer srfObstacles.Free()
	if err != nil {
		return nil, err
	}

	txrRoads, srfRoads, err := loadTexture(execPath + "/client/assets/Roads.png")
	defer srfRoads.Free()
	if err != nil {
		return nil, err
	}

	txrTowers, srfTowers, err := loadTexture(execPath + "/client/assets/Towers.png")
	defer srfTowers.Free()
	if err != nil {
		return nil, err
	}

	txrEnemies, srfEnemies, err := loadTexture(execPath + "/client/assets/Enemies.png")
	defer srfEnemies.Free()
	if err != nil {
		return nil, err
	}

	return &SDL{
		game: gm, pid: pid,

		window: w, renderer: r,
		windowWidth: width, windowHeight: height,

		selectedX: 0, selectedY: 0,
		selectedTower: 0,

		textures: textures{
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
		cl.textures.obstacles.Destroy()
		cl.textures.obstacles = nil
	}
	if cl.textures.roads != nil {
		cl.textures.roads.Destroy()
		cl.textures.roads = nil
	}
	if cl.textures.towers != nil {
		cl.textures.towers.Destroy()
		cl.textures.towers = nil
	}
	if cl.textures.enemies != nil {
		cl.textures.enemies.Destroy()
		cl.textures.enemies = nil
	}
}

func (cl *SDL) Draw(processTime time.Duration) error {
	if err := cl.renderer.SetDrawColor(0, 0, 0, 255); err != nil {
		return err
	}
	if err := cl.renderer.Clear(); err != nil {
		return err
	}

	for y := range cl.game.GC.FieldHeight {
		for x := range cl.game.GC.FieldWidth {
			if x == cl.selectedX && y == cl.selectedY {
			} else if obj := cl.game.GetCollisions(x, y); len(obj) > 0 {
				sheet, src := cl.textures.obstacles, &sdl.Rect{X: -64, Y: -64, W: 64, H: 64}
				switch o := obj[len(obj)-1]; o.Type() {
				case "Obstacle":
					sheet = cl.textures.obstacles
					switch o.(*game.ObstacleObj).Name {
					case "lake":
						src = &sdl.Rect{X: 0, Y: 0, W: 64, H: 64}
					case "sea":
						src = &sdl.Rect{X: 64, Y: 0, W: 64, H: 64}
					case "sand":
						src = &sdl.Rect{X: 128, Y: 0, W: 64, H: 64}
					case "hills":
						src = &sdl.Rect{X: 0, Y: 64, W: 64, H: 64}
					case "tree":
						src = &sdl.Rect{X: 64, Y: 64, W: 64, H: 64}
					case "brick":
						src = &sdl.Rect{X: 128, Y: 64, W: 64, H: 64}
					}

				case "Road":
					sheet = cl.textures.roads
					if o.(*game.RoadObj).Index == 0 {
						switch o.(*game.RoadObj).DirExit {
						case "right":
							src = &sdl.Rect{X: 0, Y: 128, W: 64, H: 64}
						case "down":
							src = &sdl.Rect{X: 64, Y: 128, W: 64, H: 64}
						case "left":
							src = &sdl.Rect{X: 128, Y: 128, W: 64, H: 64}
						case "up":
							src = &sdl.Rect{X: 192, Y: 128, W: 64, H: 64}
						}
					} else if o.(*game.RoadObj).Index == len(cl.game.GS.Roads)-1 {
						switch o.(*game.RoadObj).DirEntrance {
						case "left":
							src = &sdl.Rect{X: 0, Y: 192, W: 64, H: 64}
						case "up":
							src = &sdl.Rect{X: 64, Y: 192, W: 64, H: 64}
						case "right":
							src = &sdl.Rect{X: 128, Y: 192, W: 64, H: 64}
						case "down":
							src = &sdl.Rect{X: 192, Y: 192, W: 64, H: 64}
						}
					} else {
						switch o.(*game.RoadObj).DirEntrance + ";" + o.(*game.RoadObj).DirExit {
						case "up;down", "down;up":
							src = &sdl.Rect{X: 64, Y: 0, W: 64, H: 64}
						case "left;right", "right;left":
							src = &sdl.Rect{X: 128, Y: 0, W: 64, H: 64}
						case "left;down", "down;left":
							src = &sdl.Rect{X: 0, Y: 64, W: 64, H: 64}
						case "up;left", "left;up":
							src = &sdl.Rect{X: 64, Y: 64, W: 64, H: 64}
						case "up;right", "right;up":
							src = &sdl.Rect{X: 128, Y: 64, W: 64, H: 64}
						case "right;down", "down;right":
							src = &sdl.Rect{X: 192, Y: 64, W: 64, H: 64}
						}
					}

				case "Tower":
					// TODO: Tower oriantation.

					sheet = cl.textures.towers
					switch o.(*game.TowerObj).Name {
					case "Basic":
						src = &sdl.Rect{X: 0, Y: 0, W: 64, H: 64}
					case "LongRange":
						src = &sdl.Rect{X: 0, Y: 64, W: 64, H: 64}
					case "Fast":
						src = &sdl.Rect{X: 0, Y: 128, W: 64, H: 64}
					case "Strong":
						src = &sdl.Rect{X: 0, Y: 192, W: 64, H: 64}
					}

				case "Enemy":
					if o.(*game.EnemyObj).Progress < 0.5 {
						continue
					}

					sheet = cl.textures.enemies
					road := cl.game.GS.Roads[max(int(o.(*game.EnemyObj).Progress), len(cl.game.GS.Roads)-1)]
					switch road.DirEntrance + ";" + road.DirExit {
					case "up;down": //
						src = &sdl.Rect{X: 0, Y: 0, W: 64, H: 64}
					case "up;left", "right;down":
						src = &sdl.Rect{X: 0, Y: 0, W: 64, H: 64}
					case "right;left": //
						src = &sdl.Rect{X: 0, Y: 0, W: 64, H: 64}
					case "right;up", "down;left":
						src = &sdl.Rect{X: 0, Y: 0, W: 64, H: 64}
					case "down;up": //
						src = &sdl.Rect{X: 0, Y: 0, W: 64, H: 64}
					case "down;right", "left;up":
						src = &sdl.Rect{X: 0, Y: 0, W: 64, H: 64}
					case "left;right": //
						src = &sdl.Rect{X: 0, Y: 0, W: 64, H: 64}
					case "left;down", "up;right":
						src = &sdl.Rect{X: 0, Y: 0, W: 64, H: 64}
					}
				}

				if err := cl.renderer.Copy(sheet, src, &sdl.Rect{X: int32(x * 64), Y: int32(y * 64), W: 64, H: 64}); err != nil {
					return err
				}
			}
		}
	}

	cl.renderer.Present()
	return nil
}

func (cl *SDL) Input() error {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch event.(type) {
		case *sdl.QuitEvent:
			return game.Errors.Exit
		}
	}

	return nil
}
