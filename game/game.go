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
		GamePhaseStopped,
		InvalidPlacement, InvalidSelection, InvalidPlayer,
		TowerNotExists,
		InsufficientFunds,
		Exit error
	}

	GameConfig struct {
		// Valid modes: `singleplayer`, `multiplayer`, `server`
		Mode             string
		IP               string
		Port             uint16
		FieldWidth       int
		FieldHeight      int
		GameSpeed        int
		RefuntMultiplier float64
		TickDelay        time.Duration
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
		exit    chan error
	}
)

var (
	Errors = gameErrors{
		GameStateNotWaiting:  errors.New("game state is not waiting"),
		GameStateNotActive:   errors.New("game state is not started or paused"),
		GamePhaseNotBuilding: errors.New("game phase is not building"),
		GamePhaseStopped:     errors.New("game phase is stopped"),
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
			Name:           "Soldier",
			Cost:           25,
			Range:          3,
			Rotation:       0.0,
			damage:         1,
			reloadSpeed:    1.0,
			ReloadProgress: 0.0,
			effectiveRange: []*RoadObj{},
		}, {
			x: 0, y: 0, UID: -1,
			Name:           "Sniper",
			Cost:           50,
			Range:          10,
			Rotation:       0.0,
			damage:         1,
			reloadSpeed:    0.25,
			ReloadProgress: 0.0,
			effectiveRange: []*RoadObj{},
		}, {
			x: 0, y: 0, UID: -1,
			Name:           "Scout",
			Cost:           75,
			Range:          2,
			Rotation:       0.0,
			damage:         1,
			reloadSpeed:    1.5,
			ReloadProgress: 0.0,
			effectiveRange: []*RoadObj{},
		}, {
			x: 0, y: 0, UID: -1,
			ReloadProgress: 0.0, Rotation: 0.0, effectiveRange: []*RoadObj{},
			Name:        "Heavy",
			Cost:        75,
			Range:       2,
			damage:      5,
			reloadSpeed: 0.5,
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
		exit:    make(chan error),
	}
}

func (game *Game) Run(callback func(time.Duration) error) error {
	processTime := time.Duration(0)
	last := time.Now()
	for game.GS.State != "stopped" {
		now := time.Now()

		game.iterate(time.Since(last) * time.Duration(game.GC.GameSpeed))
		if err := callback(processTime); err != nil {
			if err == Errors.Exit {
				return nil
			}
			return err
		}

		last = now
		processTime = time.Since(now)
		time.Sleep(game.GC.TickDelay - time.Since(now))
	}
	return nil
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

func (game *Game) AddPlayer() int {
	index := len(game.Players)
	game.Players = append(game.Players, Player{
		Index: index,
		Coins: 80,
	})
	return index
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
	game.spawnEnemies()

	game.GS.Phase = "defending"
	return nil
}

func (game *Game) PlaceTower(name string, x, y, pid int) error {
	if game.GS.State != "started" && game.GS.State != "paused" {
		return Errors.GameStateNotActive
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

func (game *Game) DestroyTower(x, y, pid int) error {
	if game.GS.State != "started" && game.GS.State != "paused" {
		return Errors.GameStateNotActive
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

	game.Players[pid].Coins += int(float64(towers[0].Cost) * game.GC.RefuntMultiplier)
	game.GS.Towers = slices.DeleteFunc(game.GS.Towers, func(obj *TowerObj) bool { return obj.UID == towers[0].UID })

	return nil
}

func (game *Game) DestroyObstacle(x, y, pid int) error {
	if game.GS.State != "started" && game.GS.State != "paused" {
		return Errors.GameStateNotActive
	}
	if pid < 0 || pid >= len(game.Players) {
		return Errors.InvalidPlayer
	}
	obstacles := game.GetCollisionObstacles(x, y)
	if len(obstacles) != 1 {
		return Errors.InvalidSelection
	}
	obstacle := obstacles[0]

	if obstacle.Cost > game.Players[pid].Coins {
		return Errors.InsufficientFunds
	}
	game.Players[pid].Coins -= obstacle.Cost

	game.GS.Obstacles = slices.DeleteFunc(game.GS.Obstacles, func(obj *ObstacleObj) bool { return obj.UID == obstacle.UID })

	return nil
}

func (game *Game) genRoads() {
	x, y, dir := rand.IntN(game.GC.FieldWidth), rand.IntN(game.GC.FieldHeight), [4]string{"up", "right", "down", "left"}[rand.IntN(4)]
	index, conRetries := 0, 0
	for i := 0; i < int(float64(game.GC.FieldWidth+game.GC.FieldHeight)*(1+rand.Float64())); i++ {
		oldX, oldY, oldDir := x, y, dir

		switch n := rand.IntN(8); {
		case n == 0 && oldDir != "down":
			dir = "up"
		case n == 1 && oldDir != "left":
			dir = "right"
		case n == 2 && oldDir != "up":
			dir = "down"
		case n == 3 && oldDir != "right":
			dir = "left"
		default:
		}

		switch dir {
		case "up":
			y -= 1
			if y < 0 {
				y = game.GC.FieldHeight - 1
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
				x = game.GC.FieldWidth - 1
			}
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
		}

		if (conRetries < 2 && game.CheckCollisions(x, y)) || (game.CheckCollisionObstacles(x, y) || game.CheckCollisionTowers(x, y)) {
			x, y, dir = oldX, oldY, oldDir
			if conRetries < 8 && i < game.GC.FieldWidth+game.GC.FieldHeight {
				i--
			}
			conRetries++
			continue
		}
		conRetries = 0

		game.GS.Roads = append(game.GS.Roads, &RoadObj{
			x: oldX, y: oldY,
			Index:       index,
			DirEntrance: dirEntrance, DirExit: dir,
		})
		index += 1
	}
	game.GS.Roads[0].DirEntrance = "start"
	game.GS.Roads[len(game.GS.Roads)-1].DirExit = "end"
}

func (game *Game) genObstacles() {
	for range int(float64(game.GC.FieldWidth+game.GC.FieldHeight) * rand.Float64()) {
		x, y := rand.IntN(game.GC.FieldWidth), rand.IntN(game.GC.FieldHeight)

		if game.CheckCollisions(x, y) {
			continue
		}

		uid += 1
		game.GS.Obstacles = append(game.GS.Obstacles, &ObstacleObj{
			x: x, y: y,
			UID:  uid,
			Cost: 100,
		})
	}
}

func (game *Game) iterate(delta time.Duration) {
	if game.GS.State == "paused" {
		return
	}

	if game.GS.Phase == "defending" {
		for _, tower := range game.GS.Towers {
			if tower.ReloadProgress < 1 {
				tower.ReloadProgress += (float64(delta.Milliseconds()) / 1000) * tower.reloadSpeed
			}
			if tower.ReloadProgress < 1 {
				continue
			}

			for _, road := range tower.effectiveRange {
				enemies := game.GetCollisionEnemies(road.x, road.y)
				i := slices.IndexFunc(enemies, func(obj *EnemyObj) bool { return obj.startDelay <= 0 })
				if i < 0 {
					continue
				}
				enemies[i].Health -= min(enemies[i].Health, tower.damage)
				tower.ReloadProgress -= 1
				tower.Rotation = (math.Atan2(float64(enemies[i].y-tower.y), float64(enemies[i].x-tower.x)) * (180 / math.Pi)) + 90
				if tower.Rotation < 0 {
					tower.Rotation += 360
				}

				if enemies[i].Health <= 0 {
					game.Players[max(len(game.Players)-1, tower.Owner)].Coins += enemies[i].reward
					game.GS.Enemies = slices.DeleteFunc(game.GS.Enemies, func(obj *EnemyObj) bool { return obj.UID == enemies[i].UID })
				}
				break
			}
		}

		toPop := []int{}
		for i, enemy := range game.GS.Enemies {
			if enemy.startDelay > 0 {
				enemy.startDelay -= int(delta.Milliseconds())
				if enemy.startDelay < 0 {
					enemy.Progress += (float64(-enemy.startDelay) / 1000) * enemy.speedMultiplier
					enemy.startDelay = 0
				}
				continue
			}

			enemy.Progress += (float64(delta.Milliseconds()) / 1000) * enemy.speedMultiplier

			if int(enemy.Progress) >= len(game.GS.Roads) {
				game.GS.Health = max(game.GS.Health-enemy.Health, 0)
				toPop = append(toPop, i)
				continue
			}
			enemy.x, enemy.y = game.GS.Roads[int(enemy.Progress)].Cord()
		}
		slices.Reverse(toPop)
		for _, i := range toPop {
			game.GS.Enemies = slices.Delete(game.GS.Enemies, i, i+1)
		}

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
