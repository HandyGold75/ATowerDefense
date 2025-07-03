package game

import "slices"

type (
	GameObj interface {
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
		// Targeting range in tiles.
		Range int
		// Tower roation in degrees, starting north going clockwise.
		Rotation float64
		// Damage multiplier.
		damage int
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

func (game *Game) CheckCollisions(x, y int) bool {
	return game.CheckCollisionObstacles(x, y) || game.CheckCollisionRoads(x, y) || game.CheckCollisionTowers(x, y) || game.CheckCollisionEnemies(x, y)
}

func (game *Game) GetCollisions(x, y int) []GameObj {
	objects := []GameObj{}
	for _, obj := range game.GetCollisionObstacles(x, y) {
		objects = append(objects, obj)
	}
	for _, obj := range game.GetCollisionRoads(x, y) {
		objects = append(objects, obj)
	}
	for _, obj := range game.GetCollisionTowers(x, y) {
		objects = append(objects, obj)
	}
	for _, obj := range game.GetCollisionEnemies(x, y) {
		objects = append(objects, obj)
	}
	return objects
}

func (obj *ObstacleObj) Cord() (int, int) { return obj.x, obj.y }

func (game *Game) CheckCollisionObstacles(x, y int) bool {
	return slices.ContainsFunc(game.GS.Obstacles, func(obj *ObstacleObj) bool { return obj.x == x && obj.y == y })
}

func (game *Game) GetCollisionObstacles(x, y int) []*ObstacleObj {
	return slices.Collect(
		func(yield func(*ObstacleObj) bool) {
			for _, obj := range game.GS.Obstacles {
				if obj.x == x && obj.y == y {
					if !yield(obj) {
						return
					}
				}
			}
		},
	)
}

func (obj *RoadObj) Cord() (int, int) { return obj.x, obj.y }

func (game *Game) CheckCollisionRoads(x, y int) bool {
	return slices.ContainsFunc(game.GS.Roads, func(obj *RoadObj) bool { return obj.x == x && obj.y == y })
}

func (game *Game) GetCollisionRoads(x, y int) []*RoadObj {
	return slices.Collect(
		func(yield func(*RoadObj) bool) {
			for _, obj := range game.GS.Roads {
				if obj.x == x && obj.y == y {
					if !yield(obj) {
						return
					}
				}
			}
		},
	)
}

func (obj *TowerObj) Cord() (int, int) { return obj.x, obj.y }

func (game *Game) CheckCollisionTowers(x, y int) bool {
	return slices.ContainsFunc(game.GS.Towers, func(obj *TowerObj) bool { return obj.x == x && obj.y == y })
}

func (game *Game) GetCollisionTowers(x, y int) []*TowerObj {
	return slices.Collect(
		func(yield func(*TowerObj) bool) {
			for _, obj := range game.GS.Towers {
				if obj.x == x && obj.y == y {
					if !yield(obj) {
						return
					}
				}
			}
		},
	)
}

func (obj *EnemyObj) Cord() (int, int) { return obj.x, obj.y }

func (game *Game) CheckCollisionEnemies(x, y int) bool {
	return slices.ContainsFunc(game.GS.Enemies, func(obj *EnemyObj) bool { return obj.x == x && obj.y == y })
}

func (game *Game) GetCollisionEnemies(x, y int) []*EnemyObj {
	return slices.Collect(
		func(yield func(*EnemyObj) bool) {
			for _, obj := range game.GS.Enemies {
				if obj.x == x && obj.y == y {
					if !yield(obj) {
						return
					}
				}
			}
		},
	)
}
