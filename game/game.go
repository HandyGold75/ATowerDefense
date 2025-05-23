package game

import (
	"errors"
	"math/rand/v2"
	"slices"
	"time"
)

type (
	color string

	gameErrors struct{ GameStateNotWaiting, GameStateNotActive, GamePhaseNotBuilding, InvalidPlacement, TowerNotExists error }

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
		State string
		// Valid phases: `building`, `defending`, `lost`
		Phase     string
		Health    int
		Obstacles []*ObstacleObj
		Roads     []*RoadObj
		Towers    []*TowerObj
		Enemies   []*EnemyObj
	}
	Game struct {
		GC GameConfig
		GS GameState
	}
)

var (
	Errors = gameErrors{
		GameStateNotWaiting:  errors.New("game state is not waiting"),
		GameStateNotActive:   errors.New("game state is not started or paused"),
		GamePhaseNotBuilding: errors.New("game phase is not building"),
		InvalidPlacement:     errors.New("object is placed invalid"),
		TowerNotExists:       errors.New("tower does not exists"),
	}

	Towers = []TowerObj{
		{
			x: 0, y: 0,
			color:               Black,
			Name:                "Basic",
			damage:              1,
			fireRange:           3,
			fireProgress:        0.0,
			fireSpeedMultiplier: 1.0,
			effectiveRange:      []*RoadObj{},
		}, {
			x: 0, y: 0,
			color:               Black,
			Name:                "LongRange",
			damage:              1,
			fireRange:           5,
			fireProgress:        0.0,
			fireSpeedMultiplier: 1.0,
			effectiveRange:      []*RoadObj{},
		}, {
			x: 0, y: 0,
			color:               Black,
			Name:                "Fast",
			damage:              1,
			fireRange:           1,
			fireProgress:        0.0,
			fireSpeedMultiplier: 3.0,
			effectiveRange:      []*RoadObj{},
		}, {
			x: 0, y: 0,
			color:               Black,
			Name:                "Strong",
			damage:              3,
			fireRange:           1,
			fireProgress:        0.0,
			fireSpeedMultiplier: 1.0,
			effectiveRange:      []*RoadObj{},
		},
	}

	uid = 0
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

func NewGame(gc GameConfig) *Game {
	return &Game{
		GC: gc,
		GS: GameState{
			State:     "waiting",
			Phase:     "building",
			Health:    100,
			Obstacles: []*ObstacleObj{},
			Roads:     []*RoadObj{},
			Towers:    []*TowerObj{},
			Enemies:   []*EnemyObj{},
		},
	}
}

func (game *Game) genRoads() {
	x, y, dir := rand.IntN(game.GC.FieldWidth), rand.IntN(game.GC.FieldHeight), "right"
	index := 0
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
			y -= 1
			if y < 0 {
				y = game.GC.FieldHeight
			}
		case "right":
			x += 1
			if x >= game.GC.FieldWidth {
				x = 0
			}
		case "down":
			y += 1
			if y >= game.GC.FieldHeight {
				y = 0
			}
		case "left":
			x -= 1
			if x < 0 {
				x = game.GC.FieldWidth
			}
		}

		if game.CheckCollisionObstacles(x, y) || game.CheckCollisionTowers(x, y) {
			x, y, dir = oldX, oldY, oldDir
			continue
		}

		game.GS.Roads = append(game.GS.Roads, &RoadObj{
			x: oldX, y: oldY,
			color:     BGGreen + White,
			Index:     index,
			Direction: dir,
		})
		index += 1
	}
}

func (game *Game) genObstacles() {
	for range int(float64(game.GC.FieldWidth+game.GC.FieldHeight) * rand.Float64()) {
		x, y := rand.IntN(game.GC.FieldWidth), rand.IntN(game.GC.FieldHeight)

		if game.CheckCollisions(x, y) {
			continue
		}

		game.GS.Obstacles = append(game.GS.Obstacles, &ObstacleObj{
			x: x, y: y,
			color: BGBrightYellow + BrightBlue,
		})
	}
}

func (game *Game) genEnemies() {
	x, y := 0, 0
	if len(game.GS.Roads) > 0 {
		x, y = game.GS.Roads[0].Cord()
	}

	for i := range 10 + rand.IntN(10) {
		uid += 1
		game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
			x: x, y: y,
			color:           BGGreen + Red,
			UID:             uid,
			Progress:        0.0,
			health:          1,
			startDelay:      i * 1500,
			speedMultiplier: 4.0,
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

func (game *Game) StartRound() error {
	if game.GS.State != "started" && game.GS.State != "paused" {
		return Errors.GameStateNotActive
	} else if game.GS.Phase != "building" {
		return Errors.GamePhaseNotBuilding
	}

	game.genEnemies()

	// TEMP
	for range int(float64(game.GC.FieldWidth+game.GC.FieldHeight) * rand.Float64()) {
		_ = game.PlaceTower("", rand.IntN(game.GC.FieldWidth), rand.IntN(game.GC.FieldHeight))
	}

	game.GS.Phase = "defending"
	return nil
}

func (game *Game) PlaceTower(name string, x, y int) error {
	if game.GS.State != "started" && game.GS.State != "paused" {
		return Errors.GameStateNotActive
	} else if game.GS.Phase != "building" {
		return Errors.GamePhaseNotBuilding
	}
	if game.CheckCollisions(x, y) {
		return Errors.InvalidPlacement
	}

	i := slices.IndexFunc(Towers, func(obj TowerObj) bool { return obj.x == x && obj.y == y })
	if i < 0 {
		return Errors.TowerNotExists
	}
	tower := Towers[i]

	for offsetY := range (tower.fireRange * 2) + 1 {
		for offsetX := range (tower.fireRange * 2) + 1 {
			tower.effectiveRange = append(tower.effectiveRange, game.GetCollisionRoads(x+(offsetX-tower.fireRange), y+(offsetY-tower.fireRange))...)
		}
	}
	slices.SortFunc(tower.effectiveRange, func(a, b *RoadObj) int { return b.Index - a.Index })

	game.GS.Towers = append(game.GS.Towers, &tower)

	return nil
}

func (game *Game) iterateTowers(delta time.Duration) {
	for _, tower := range game.GS.Towers {
		if tower.fireProgress < 1 {
			tower.fireProgress += (float64(delta.Milliseconds()) / 1000) * tower.fireSpeedMultiplier
		}
		if tower.fireProgress < 1 {
			continue
		}

		for _, road := range tower.effectiveRange {
			enemies := game.GetCollisionEnemies(road.x, road.y)
			i := slices.IndexFunc(enemies, func(obj *EnemyObj) bool { return obj.startDelay > 0 })
			if i < 0 {
				continue
			}
			enemies[i].health -= min(enemies[0].health, tower.damage)
			tower.fireProgress -= 1

			if enemies[i].health <= 0 {
				game.GS.Enemies = slices.DeleteFunc(game.GS.Enemies, func(obj *EnemyObj) bool { return obj.UID == enemies[i].UID })
			}
			break
		}
	}
}

func (game *Game) iterateEnemies(delta time.Duration) {
	toPop := []int{}
	for i, enemy := range game.GS.Enemies {
		if enemy.startDelay > 0 {
			enemy.startDelay -= min(enemy.startDelay, int(delta.Milliseconds()))
			continue
		}

		enemy.Progress += (float64(delta.Milliseconds()) / 1000) * enemy.speedMultiplier

		if int(enemy.Progress) >= len(game.GS.Roads) {
			game.GS.Health -= 1
			toPop = append(toPop, i)
			continue
		}

		enemy.x, enemy.y = game.GS.Roads[int(enemy.Progress)].Cord()
	}
	slices.Reverse(toPop)
	for _, i := range toPop {
		game.GS.Enemies = slices.Delete(game.GS.Enemies, i, i+1)
	}
}

func (game *Game) Iterate(delta time.Duration) {
	if game.GS.State == "paused" {
		return
	}

	if game.GS.Phase == "defending" {
		game.iterateTowers(delta)
		game.iterateEnemies(delta)

		if len(game.GS.Enemies) <= 0 {
			game.GS.Phase = "building"
		}
		if game.GS.Health <= 0 {
			game.GS.Phase = "lost"
			return
		}
	}
}
