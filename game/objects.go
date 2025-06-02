package game

import "slices"

type (
	GameObj interface {
		// Valid types: `Obstacle`, `Road`, `Tower`, `Enemy`.
		Type() string
		// X, Y
		Cord() (int, int)
	}

	ObstacleObj struct {
		x, y int
		Name string
	}
	RoadObj struct {
		x, y int
		// Road index
		Index int
		// Valid directions: `up`, `right`, `down`, `left`.
		DirEntrance, DirExit string
	}
	TowerObj struct {
		x, y int
		// Unique identifier
		UID int
		// Tower name
		Name string
		// Cost of the tower.
		Cost int
		// Owner of the tower by player index.
		Owner int
		// Damage multiplier.
		damage int
		// Targeting range in tiles.
		fireRange int
		// Fire when progress hits 1
		fireProgress float64
		// Progress 1 every second * this.
		fireSpeedMultiplier float64
		// Road objects the tower has range over.
		effectiveRange []*RoadObj
	}
	EnemyObj struct {
		x, y int
		// Unique identifier
		UID int
		// Every 1 progress represents 1 tile moved.
		Progress float64
		// Amount of coins given once defeated.
		reward int
		// Despawn when <= 0.
		health int
		// Delay spawning by this compared to phase start in ms.
		startDelay int
		// Progress 1 every second * this.
		speedMultiplier float64
	}
)

func (obj *ObstacleObj) Type() string     { return "Obstacle" }
func (obj *ObstacleObj) Cord() (int, int) { return obj.x, obj.y }

func (game *Game) CheckCollisionObstacles(x, y int) bool {
	return slices.ContainsFunc(game.GS.Obstacles, func(obj *ObstacleObj) bool { return obj.x == x && obj.y == y })
}

func (game *Game) GetCollisionObstacles(x, y int) []*ObstacleObj {
	objects := []*ObstacleObj{}
	if i := slices.IndexFunc(game.GS.Obstacles, func(obj *ObstacleObj) bool { return obj.x == x && obj.y == y }); i >= 0 {
		objects = append(objects, game.GS.Obstacles[i])
	}
	return objects
}

func (obj *RoadObj) Type() string     { return "Road" }
func (obj *RoadObj) Cord() (int, int) { return obj.x, obj.y }

func (game *Game) CheckCollisionRoads(x, y int) bool {
	return slices.ContainsFunc(game.GS.Roads, func(obj *RoadObj) bool { return obj.x == x && obj.y == y })
}

func (game *Game) GetCollisionRoads(x, y int) []*RoadObj {
	objects := []*RoadObj{}
	if i := slices.IndexFunc(game.GS.Roads, func(obj *RoadObj) bool { return obj.x == x && obj.y == y }); i >= 0 {
		objects = append(objects, game.GS.Roads[i])
	}
	return objects
}

func (obj *TowerObj) Type() string     { return "Tower" }
func (obj *TowerObj) Cord() (int, int) { return obj.x, obj.y }

func (game *Game) CheckCollisionTowers(x, y int) bool {
	return slices.ContainsFunc(game.GS.Towers, func(obj *TowerObj) bool { return obj.x == x && obj.y == y })
}

func (game *Game) GetCollisionTowers(x, y int) []*TowerObj {
	objects := []*TowerObj{}
	if i := slices.IndexFunc(game.GS.Towers, func(obj *TowerObj) bool { return obj.x == x && obj.y == y }); i >= 0 {
		objects = append(objects, game.GS.Towers[i])
	}
	return objects
}

func (obj *EnemyObj) Type() string     { return "Enemy" }
func (obj *EnemyObj) Cord() (int, int) { return obj.x, obj.y }

func (game *Game) CheckCollisionEnemies(x, y int) bool {
	return slices.ContainsFunc(game.GS.Enemies, func(obj *EnemyObj) bool { return obj.x == x && obj.y == y })
}

func (game *Game) GetCollisionEnemies(x, y int) []*EnemyObj {
	objects := []*EnemyObj{}
	if i := slices.IndexFunc(game.GS.Enemies, func(obj *EnemyObj) bool { return obj.x == x && obj.y == y }); i >= 0 {
		objects = append(objects, game.GS.Enemies[i])
	}
	return objects
}

func (game *Game) CheckCollisions(x, y int) bool {
	return game.CheckCollisionObstacles(x, y) || game.CheckCollisionRoads(x, y) || game.CheckCollisionTowers(x, y) || game.CheckCollisionEnemies(x, y)
}

func (game *Game) GetCollisions(x, y int) []GameObj {
	objects := []GameObj{}
	if i := slices.IndexFunc(game.GS.Obstacles, func(obj *ObstacleObj) bool { return obj.x == x && obj.y == y }); i >= 0 {
		objects = append(objects, game.GS.Obstacles[i])
	}
	if i := slices.IndexFunc(game.GS.Roads, func(obj *RoadObj) bool { return obj.x == x && obj.y == y }); i >= 0 {
		objects = append(objects, game.GS.Roads[i])
	}
	if i := slices.IndexFunc(game.GS.Towers, func(obj *TowerObj) bool { return obj.x == x && obj.y == y }); i >= 0 {
		objects = append(objects, game.GS.Towers[i])
	}
	if i := slices.IndexFunc(game.GS.Enemies, func(obj *EnemyObj) bool { return obj.x == x && obj.y == y }); i >= 0 {
		objects = append(objects, game.GS.Enemies[i])
	}
	return objects
}
