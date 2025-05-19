package game

import (
	"errors"
	"math/rand/v2"
)

type (
	errGame struct {
		GameStateNotWaiting error
		GameStateNotActive  error
	}

	keyBinds struct {
		ESC, P,
		CTRL_C, CTRL_D, Q,
		W, D, S, A, K, L, J, H, UP, RIGHT, DOWN, LEFT []byte
	}

	GameConfig struct {
		// Valid modes: `singleplayer`, `multiplayer`, `server`
		Mode        string
		IP          string
		Port        uint16
		FieldHeight int
		FieldWidth  int
	}

	FieldObj struct {
		x, y  int
		color string
	}
	RoadObj struct {
		x, y  int
		color string
	}
	TowerObj struct {
		x, y  int
		color string
	}
	EnemieObj struct {
		x, y  int
		color string
	}
	GameState struct {
		// Valid states: `waiting`, `started`, `paused`, `stopped`
		State   string
		field   []FieldObj
		road    []RoadObj
		towers  []TowerObj
		enemies []EnemieObj
	}
	Game struct {
		KeyBinds  keyBinds
		GC        GameConfig
		GS        GameState
		Iteration uint8
	}
)

var ErrGame = errGame{
	GameStateNotWaiting: errors.New("game state is not waiting"),
	GameStateNotActive:  errors.New("game state is not started or paused"),
}

const (
	Reset string = "\033[0m"

	Black   string = "\033[30m"
	Red     string = "\033[31m"
	Green   string = "\033[32m"
	Yellow  string = "\033[33m"
	Blue    string = "\033[34m"
	Magenta string = "\033[35m"
	Cyan    string = "\033[36m"
	White   string = "\033[37m"
)

func NewGame(gc GameConfig) *Game {
	return &Game{
		KeyBinds: keyBinds{
			ESC: []byte{27, 0, 0}, P: []byte{112, 0, 0},
			CTRL_C: []byte{3, 0, 0}, CTRL_D: []byte{4, 0, 0}, Q: []byte{113, 0, 0},
			W: []byte{119, 0, 0}, D: []byte{100, 0, 0}, S: []byte{115, 0, 0}, A: []byte{97, 0, 0},
			K: []byte{107, 0, 0}, L: []byte{108, 0, 0}, J: []byte{106, 0, 0}, H: []byte{104, 0, 0},
			UP: []byte{27, 91, 65}, RIGHT: []byte{27, 91, 67}, DOWN: []byte{27, 91, 66}, LEFT: []byte{27, 91, 68},
		},
		GC: gc,
		GS: GameState{
			State:   "waiting",
			field:   []FieldObj{},
			road:    []RoadObj{},
			towers:  []TowerObj{},
			enemies: []EnemieObj{},
		},
	}
}

func (game *Game) Start() error {
	if game.GS.State != "waiting" {
		return ErrGame.GameStateNotWaiting
	}

	x, y := 0, 0
	for range int(float64(game.GC.FieldHeight+game.GC.FieldWidth) / 10) {
		switch rand.IntN(10) {
		case 0:
		case 1:
		case 2:
		case 3:
		default:
		}
	}

	game.GS.State = "started"
	return nil
}

func (game *Game) Stop() error {
	if game.GS.State != "started" || game.GS.State != "paused" {
		return ErrGame.GameStateNotActive
	}

	game.GS.State = "stopped"
	return nil
}

func (game *Game) TogglePause() error {
	if game.GS.State == "started" {
		game.GS.State = "paused"
		return nil
	} else if game.GS.State != "paused" {
		game.GS.State = "started"
		return nil
	}
	return ErrGame.GameStateNotActive
}

func (game *Game) Iterate() {
	if game.GS.State == "paused" {
		return
	}
	game.Iteration += 1
}
