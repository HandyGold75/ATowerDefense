package game

import (
	"errors"
	"math"
	"math/rand/v2"
	"slices"
	"time"
)

type (
	gameErrors struct {
		GameStateNotWaiting, GameStateNotActive, GamePhaseNotBuilding,
		InvalidPlacement, InvalidSelection, InvalidPlayer,
		TowerNotExists,
		InsufficientFunds,
		Exit error
	}

	GameConfig struct {
		// Valid modes: `singleplayer`, `multiplayer`, `server`
		Mode        string
		IP          string
		Port        uint16
		FieldHeight int
		FieldWidth  int
		TickDelay   time.Duration
	}
	GameState struct {
		// Valid states: `waiting`, `started`, `paused`, `stopped`
		State string
		// Valid phases: `building`, `defending`, `lost`
		Phase     string
		Round     int
		Health    int
		Obstacles []*ObstacleObj
		Roads     []*RoadObj
		Towers    []*TowerObj
		Enemies   []*EnemyObj
	}
	Player struct {
		Index int
		Coins int
	}
	Game struct {
		GC      GameConfig
		GS      GameState
		Players []Player
	}
)

var (
	Errors = gameErrors{
		GameStateNotWaiting:  errors.New("game state is not waiting"),
		GameStateNotActive:   errors.New("game state is not started or paused"),
		GamePhaseNotBuilding: errors.New("game phase is not building"),
		InvalidPlacement:     errors.New("object is placed invalid"),
		InvalidSelection:     errors.New("selection is invalid"),
		InvalidPlayer:        errors.New("player is invalid"),
		TowerNotExists:       errors.New("tower does not exists"),
		InsufficientFunds:    errors.New("not enough funds"),
		Exit:                 errors.New("game is exiting"),
	}

	Towers = []TowerObj{
		{
			x: 0, y: 0, UID: -1,
			Name:                "Basic",
			Cost:                25,
			Range:               3,
			Rotation:            0.0,
			damage:              1,
			fireProgress:        0.0,
			fireSpeedMultiplier: 1.0,
			effectiveRange:      []*RoadObj{},
		}, {
			x: 0, y: 0, UID: -1,
			Name:                "LongRange",
			Cost:                30,
			Range:               6,
			Rotation:            0.0,
			damage:              1,
			fireProgress:        0.0,
			fireSpeedMultiplier: 0.75,
			effectiveRange:      []*RoadObj{},
		}, {
			x: 0, y: 0, UID: -1,
			Name:                "Fast",
			Cost:                40,
			Range:               2,
			Rotation:            0.0,
			damage:              1,
			fireProgress:        0.0,
			fireSpeedMultiplier: 1.75,
			effectiveRange:      []*RoadObj{},
		}, {
			x: 0, y: 0, UID: -1,
			Name:                "Strong",
			Cost:                50,
			Range:               2,
			Rotation:            0.0,
			damage:              3,
			fireProgress:        0.0,
			fireSpeedMultiplier: 0.75,
			effectiveRange:      []*RoadObj{},
		},
	}

	uid = 0
)

func NewGame(gc GameConfig) *Game {
	return &Game{
		GC: gc,
		GS: GameState{
			State:     "waiting",
			Phase:     "building",
			Round:     0,
			Health:    100,
			Obstacles: []*ObstacleObj{},
			Roads:     []*RoadObj{},
			Towers:    []*TowerObj{},
			Enemies:   []*EnemyObj{},
		},
		Players: []Player{},
	}
}

func (game *Game) genRoads() {
	x, y, dir := rand.IntN(game.GC.FieldWidth), rand.IntN(game.GC.FieldHeight), "right"
	index := 0
	for range int(float64(game.GC.FieldWidth+game.GC.FieldHeight) * (1 + rand.Float64())) {
		oldX, oldY, oldDir := x, y, dir

		switch i := rand.IntN(8); {
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
		default:
			x, y, dir = oldX, oldY, oldDir
			continue
		}

		dirEntrance := ""
		switch oldDir {
		case "up":
			dirEntrance = "down"
		case "right":
			dirEntrance = "left"
		case "down":
			dirEntrance = "up"
		case "left":
			dirEntrance = "right"
		default:
			x, y, dir = oldX, oldY, oldDir
			continue

		}

		if game.CheckCollisionObstacles(x, y) || game.CheckCollisionTowers(x, y) {
			x, y, dir = oldX, oldY, oldDir
			continue
		}

		game.GS.Roads = append(game.GS.Roads, &RoadObj{
			x: oldX, y: oldY,
			Index:       index,
			DirEntrance: dirEntrance, DirExit: dir,
		})
		index += 1
	}
}

func (game *Game) genObstacles() {
	obstaclesNames := []string{"lake", "sea", "sand", "hills", "tree", "brick"}
	for range int(float64(game.GC.FieldWidth+game.GC.FieldHeight) * rand.Float64()) {
		x, y := rand.IntN(game.GC.FieldWidth), rand.IntN(game.GC.FieldHeight)

		if game.CheckCollisions(x, y) {
			continue
		}

		game.GS.Obstacles = append(game.GS.Obstacles, &ObstacleObj{
			x: x, y: y,
			Name: obstaclesNames[rand.IntN(len(obstaclesNames))],
		})
	}
}

func (game *Game) genEnemies() {
	x, y := 0, 0
	if len(game.GS.Roads) > 0 {
		x, y = game.GS.Roads[0].Cord()
	}

	switch r := game.GS.Round; {
	case r <= 2:
		for i := range 15 * r {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y,
				UID:             uid,
				Progress:        0.0,
				reward:          1,
				health:          1,
				startDelay:      i * 1000,
				speedMultiplier: 1.0,
			})
		}
	case r <= 4:
		for i := range 10 * r {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y,
				UID:             uid,
				Progress:        0.0,
				reward:          1,
				health:          1,
				startDelay:      i * 250,
				speedMultiplier: 2.0,
			})
		}
	case r <= 6:
		for i := range 5 * r {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y,
				UID:             uid,
				Progress:        0.0,
				reward:          3,
				health:          10,
				startDelay:      i * 1000,
				speedMultiplier: 0.5,
			})
		}
	case r <= 8:
		for i := range 5 * r {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y,
				UID:             uid,
				Progress:        0.0,
				reward:          2,
				health:          5,
				startDelay:      i * 250,
				speedMultiplier: 1.0,
			})
		}
	case r <= 10:
		for i := range 10 * r {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y,
				UID:             uid,
				Progress:        0.0,
				reward:          3,
				health:          10,
				startDelay:      i * 250,
				speedMultiplier: 2.0,
			})
		}
	default:
		for i := range 15 * r {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y,
				UID:             uid,
				Progress:        0.0,
				reward:          int(float64(r) / 10),
				health:          r,
				startDelay:      i * max(10, 250-(r*10)),
				speedMultiplier: 1.0 + (float64(r) / 10),
			})
		}
	}
}

func (game *Game) AddPlayer() int {
	index := len(game.Players)
	game.Players = append(game.Players, Player{
		Index: index,
		Coins: 80,
	})
	return index
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

func (game *Game) TogglePause() {
	switch game.GS.State {
	case "started":
		game.GS.State = "paused"
	case "paused":
		game.GS.State = "started"
	}
}

func (game *Game) StartRound() error {
	if game.GS.State != "started" && game.GS.State != "paused" {
		return Errors.GameStateNotActive
	} else if game.GS.Phase != "building" {
		return Errors.GamePhaseNotBuilding
	}

	game.GS.Round += 1
	game.genEnemies()

	game.GS.Phase = "defending"
	return nil
}

func (game *Game) PlaceTower(name string, x, y, pid int) error {
	if game.GS.State != "started" && game.GS.State != "paused" {
		return Errors.GameStateNotActive
	} else if game.GS.Phase != "building" {
		return Errors.GamePhaseNotBuilding
	}
	if pid < 0 || pid >= len(game.Players) {
		return Errors.InvalidPlayer
	}
	if game.CheckCollisions(x, y) {
		return Errors.InvalidPlacement
	}

	i := slices.IndexFunc(Towers, func(obj TowerObj) bool { return obj.Name == name })
	if i < 0 {
		return Errors.TowerNotExists
	}
	tower := Towers[i]

	if tower.Cost > game.Players[pid].Coins {
		return Errors.InsufficientFunds
	}
	game.Players[pid].Coins -= tower.Cost

	uid += 1
	tower.x, tower.y, tower.UID, tower.Owner = x, y, uid, pid
	for offsetY := range (tower.Range * 2) + 1 {
		for offsetX := range (tower.Range * 2) + 1 {
			tower.effectiveRange = append(tower.effectiveRange, game.GetCollisionRoads(x+(offsetX-tower.Range), y+(offsetY-tower.Range))...)
		}
	}
	slices.SortFunc(tower.effectiveRange, func(a, b *RoadObj) int { return b.Index - a.Index })

	game.GS.Towers = append(game.GS.Towers, &tower)

	return nil
}

func (game *Game) DestoryTower(x, y, pid int) error {
	if game.GS.State != "started" && game.GS.State != "paused" {
		return Errors.GameStateNotActive
	} else if game.GS.Phase != "building" {
		return Errors.GamePhaseNotBuilding
	}
	if pid < 0 || pid >= len(game.Players) {
		return Errors.InvalidPlayer
	}
	towers := game.GetCollisionTowers(x, y)
	if len(towers) != 1 {
		return Errors.InvalidSelection
	}
	if towers[0].Owner != pid {
		return Errors.InvalidPlayer
	}

	game.Players[pid].Coins += towers[0].Cost / 2
	game.GS.Towers = slices.DeleteFunc(game.GS.Towers, func(obj *TowerObj) bool { return obj.UID == towers[0].UID })

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
			i := slices.IndexFunc(enemies, func(obj *EnemyObj) bool { return obj.startDelay <= 0 })
			if i < 0 {
				continue
			}
			enemies[i].health -= min(enemies[i].health, tower.damage)
			tower.fireProgress -= 1
			tower.Rotation = (math.Atan2(float64(enemies[i].y-tower.y), float64(enemies[i].x-tower.x)) * (180 / math.Pi)) + 90
			if tower.Rotation < 0 {
				tower.Rotation += 360
			}

			if enemies[i].health <= 0 {
				game.Players[max(len(game.Players)-1, tower.Owner)].Coins += enemies[i].reward
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
			game.GS.Health = max(game.GS.Health-1, 0)
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
		if game.GS.Health <= 0 || len(game.Players) <= 0 {
			game.GS.Round = max(game.GS.Round-1, 0)
			game.GS.Phase = "lost"
			return
		}
	}
}
