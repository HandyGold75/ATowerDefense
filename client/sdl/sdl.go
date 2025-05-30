package clsdl

import (
	"ATowerDefense/game"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

type (
	assets struct{}

	SDL struct {
		game *game.Game
		pid  int

		windowWidth, windowHeight int32

		window   *sdl.Window
		renderer *sdl.Renderer

		assets assets
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

	return &SDL{
		game: gm, pid: pid,

		windowWidth: width, windowHeight: height,

		window: w, renderer: r,
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
}

func (cl *SDL) Draw(processTime time.Duration) error {
	cl.renderer.SetDrawColor(0, 0, 0, 255)
	cl.renderer.Clear()

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
