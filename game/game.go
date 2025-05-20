package game

import (
	"errors"
	"math/rand/v2"
	"slices"
)

type (
	color   string
	charSet string

	gameErrors struct {
		NotATerm, GameStateNotWaiting, GameStateNotActive, Exit error
	}

	keybinds struct{ Up, Down, Right, Left, PanUp, PanDown, PanRight, PanLeft, Plus, Minus, Exit, Pause, Confirm, Delete, Numbers []keybind }
	keybind  []byte

	GameConfig struct {
		// Valid modes: `singleplayer`, `multiplayer`, `server`
		Mode        string
		IP          string
		Port        uint16
		FieldHeight int
		FieldWidth  int
	}
	GameState struct {
		// Valid states: `waiting`, `started`, `paused`, `stopped`
		State     string
		Obstacles []FieldObj
		Roads     []RoadObj
		Towers    []TowerObj
	}
	Game struct {
		GC        GameConfig
		GS        GameState
		Iteration uint8
	}
)

var (
	Errors = gameErrors{
		NotATerm:            errors.New("stdin/ stdout should be a terminal"),
		GameStateNotWaiting: errors.New("game state is not waiting"),
		GameStateNotActive:  errors.New("game state is not started or paused"),
		Exit:                errors.New("game is exiting"),
	}

	KeyBinds = keybinds{
		// W, K
		Up: []keybind{{119, 0, 0}, {107, 0, 0}},
		// S, J
		Down: []keybind{{115, 0, 0}, {106, 0, 0}},
		// D, L
		Right: []keybind{{100, 0, 0}, {108, 0, 0}},
		// A, H
		Left: []keybind{{97, 0, 0}, {104, 0, 0}},

		// UP
		PanUp: []keybind{{27, 91, 65}},
		// DOWN
		PanDown: []keybind{{27, 91, 66}},
		// RIGHT,
		PanRight: []keybind{{27, 91, 67}},
		// LEFT,
		PanLeft: []keybind{{27, 91, 68}},

		// PLUS
		Plus: []keybind{{43, 0, 0}},
		// MINUS
		Minus: []keybind{{45, 0, 0}},

		// ESC, CTRL_C, CTRL_D,
		Exit: []keybind{{27, 0, 0}, {3, 0, 0}, {4, 0, 0}},
		// P, Q
		Pause: []keybind{{112, 0, 0}, {113, 0, 0}},
		// RETURN
		Confirm: []keybind{{13, 0, 0}},
		// BACKSPACE, DEL
		Delete: []keybind{{127, 0, 0}, {27, 91, 51}},

		// 0, 1, 2, 3, 4, 5, 6, 7, 8, 9
		Numbers: []keybind{{48, 0, 0}, {49, 0, 0}, {50, 0, 0}, {51, 0, 0}, {52, 0, 0}, {53, 0, 0}, {54, 0, 0}, {55, 0, 0}, {56, 0, 0}, {57, 0, 0}},
	}
)

const (
	Reset color = "\033[0m"

	Bold            color = "\033[1m"
	Faint           color = "\033[2m"
	Italic          color = "\033[3m"
	Underline       color = "\033[4m"
	StrikeTrough    color = "\033[9m"
	DubbleUnderline color = "\033[21m"

	Black   color = "\033[30m"
	Red     color = "\033[31m"
	Green   color = "\033[32m"
	Yellow  color = "\033[33m"
	Blue    color = "\033[34m"
	Magenta color = "\033[35m"
	Cyan    color = "\033[36m"
	White   color = "\033[37m"

	BGBlack   color = "\033[40m"
	BGRed     color = "\033[41m"
	BGGreen   color = "\033[42m"
	BGYellow  color = "\033[43m"
	BGBlue    color = "\033[44m"
	BGMagenta color = "\033[45m"
	BGCyan    color = "\033[46m"
	BGWhite   color = "\033[47m"

	BrightBlack   color = "\033[90m"
	BrightRed     color = "\033[91m"
	BrightGreen   color = "\033[92m"
	BrightYellow  color = "\033[93m"
	BrightBlue    color = "\033[94m"
	BrightMagenta color = "\033[95m"
	BrightCyan    color = "\033[96m"
	BrightWhite   color = "\033[97m"

	BGBrightBlack   color = "\033[100m"
	BGBrightRed     color = "\033[101m"
	BGBrightGreen   color = "\033[102m"
	BGBrightYellow  color = "\033[103m"
	BGBrightBlue    color = "\033[104m"
	BGBrightMagenta color = "\033[105m"
	BGBrightCyan    color = "\033[106m"
	BGBrightWhite   color = "\033[107m"
)

const (
	Letters        charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Digits         charSet = "0123456789"
	Hex            charSet = "0123456789abcdefABCDEF"
	WhiteSpace     charSet = " "
	Punctuation    charSet = "!\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
	GeneralCharSet charSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789 !\"#$%&'()*+,-./:;<=>?@[\\]^_`{|}~"
)

func KeyBindContains(kb []keybind, b []byte) bool {
	return slices.ContainsFunc(kb, func(v keybind) bool { return slices.Equal(v, b) })
}

func KeyBindIndex(kb []keybind, b []byte) int {
	return slices.IndexFunc(kb, func(v keybind) bool { return slices.Equal(v, b) })
}

func NewGame(gc GameConfig) *Game {
	return &Game{
		GC: gc,
		GS: GameState{
			State:     "waiting",
			Obstacles: []FieldObj{},
			Roads:     []RoadObj{},
			Towers:    []TowerObj{},
		},
	}
}

func (game *Game) CheckCollisions(x, y int) bool {
	if slices.ContainsFunc(game.GS.Obstacles, func(obj FieldObj) bool { return obj.x == x && obj.y == y }) {
		return true
	}
	if slices.ContainsFunc(game.GS.Roads, func(obj RoadObj) bool { return obj.x == x && obj.y == y }) {
		return true
	}
	if slices.ContainsFunc(game.GS.Towers, func(obj TowerObj) bool { return obj.x == x && obj.y == y }) {
		return true
	}
	return false
}

func (game *Game) GetCollisions(x, y int) []GameObj {
	objects := []GameObj{}
	if i := slices.IndexFunc(game.GS.Obstacles, func(obj FieldObj) bool { return obj.x == x && obj.y == y }); i >= 0 {
		objects = append(objects, &game.GS.Obstacles[i])
	}
	if i := slices.IndexFunc(game.GS.Roads, func(obj RoadObj) bool { return obj.x == x && obj.y == y }); i >= 0 {
		objects = append(objects, &game.GS.Roads[i])
	}
	if i := slices.IndexFunc(game.GS.Towers, func(obj TowerObj) bool { return obj.x == x && obj.y == y }); i >= 0 {
		objects = append(objects, &game.GS.Towers[i])
	}
	return objects
}

func (game *Game) genRoads() {
	x, y, dir := rand.IntN(game.GC.FieldWidth), rand.IntN(game.GC.FieldHeight), "right"
	for range int(float64(game.GC.FieldWidth+game.GC.FieldHeight) * (1 + rand.Float64())) {
		oldX, oldY, oldDir := x, y, dir

		switch i := rand.IntN(6); {
		case i == 0 && dir != "down":
			dir = "up"
		case i == 1 && dir != "left":
			dir = "right"
		case i == 2 && dir != "up":
			dir = "down"
		case i == 3 && dir != "right":
			dir = "left"
		default:
		}

		switch dir {
		case "up":
			y += 1
			if y >= game.GC.FieldHeight {
				y = 0
			}
		case "right":
			x += 1
			if x >= game.GC.FieldWidth {
				x = 0
			}
		case "down":
			y -= 1
			if y < 0 {
				y = game.GC.FieldHeight
			}
		case "left":
			x -= 1
			if x < 0 {
				x = game.GC.FieldWidth
			}
		}

		if slices.ContainsFunc(game.GetCollisions(x, y), func(obj GameObj) bool { return obj.Type() != "Road" }) {
			x, y, dir = oldX, oldY, oldDir
			continue
		}

		game.GS.Roads = append(game.GS.Roads, RoadObj{
			x: x, y: y,
			color: White,
		})
	}
}

func (game *Game) genObstacles() {
	for range int(float64(game.GC.FieldWidth+game.GC.FieldHeight) * rand.Float64()) {
		x, y := rand.IntN(game.GC.FieldWidth), rand.IntN(game.GC.FieldHeight)

		if game.CheckCollisions(x, y) {
			continue
		}

		game.GS.Obstacles = append(game.GS.Obstacles, FieldObj{
			x: x, y: y,
			color: Black,
		})
	}
}

func (game *Game) Start() error {
	if game.GS.State != "waiting" {
		return Errors.GameStateNotWaiting
	}

	game.genRoads()
	game.genObstacles()

	game.GS.State = "started"
	return nil
}

func (game *Game) Stop() error {
	if game.GS.State != "started" && game.GS.State != "paused" {
		return Errors.GameStateNotActive
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
	return Errors.GameStateNotActive
}

func (game *Game) Iterate() {
	if game.GS.State == "paused" {
		return
	}
	game.Iteration += 1
}
