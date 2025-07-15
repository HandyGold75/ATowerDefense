package game

import "math/rand/v2"

func (game *Game) spawnEnemies() {
	x, y := 0, 0
	if len(game.GS.Roads) > 0 {
		x, y = game.GS.Roads[0].Cord()
	}

	switch r := game.GS.Round; {
	case r <= 1:
		for i := range 5 {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: 1, StartHealth: 1,
				reward:          1,
				startDelay:      i * 1000,
				speedMultiplier: 1,
			})
		}

	case r <= 2:
		for i := range 10 {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: 1, StartHealth: 1,
				reward:          1,
				startDelay:      i * 1000,
				speedMultiplier: 0.75,
			})
		}

	case r <= 3:
		for i := range 5 {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: 2, StartHealth: 2,
				reward:          2,
				startDelay:      i * 1500,
				speedMultiplier: 0.75,
			})
		}

	case r <= 4:
		for i := range 5 {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: 3, StartHealth: 3,
				reward:          2,
				startDelay:      i * 1500,
				speedMultiplier: 0.75,
			})
		}

	case r <= 5:
		for i := range 10 {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: 5, StartHealth: 5,
				reward:          3,
				startDelay:      i * 1500,
				speedMultiplier: 0.75,
			})
		}

	case r <= 6:
		for i := range 15 {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: 1, StartHealth: 1,
				reward:          1,
				startDelay:      i * 1000,
				speedMultiplier: 1,
			})
		}

	case r <= 7:
		for i := range 10 {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: 1, StartHealth: 1,
				reward:          1,
				startDelay:      i * 750,
				speedMultiplier: 1,
			})
		}

	case r <= 8:
		for i := range 15 {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: 1, StartHealth: 1,
				reward:          2,
				startDelay:      i * 500,
				speedMultiplier: 1.25,
			})
		}

	case r <= 9:
		for i := range 15 {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: 1, StartHealth: 1,
				reward:          2,
				startDelay:      i * 500,
				speedMultiplier: 1.5,
			})
		}

	case r <= 10:
		for i := range 30 {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: 1, StartHealth: 1,
				reward:          3,
				startDelay:      i * 250,
				speedMultiplier: 2.0,
			})
		}

	case r <= 11:
		for i := range 15 {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: 1, StartHealth: 1,
				reward:          1,
				startDelay:      i * 1000,
				speedMultiplier: 1.0,
			})
		}

	case r <= 12:
		for i := range 10 {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: 5, StartHealth: 5,
				reward:          2,
				startDelay:      i * 500,
				speedMultiplier: 1.0,
			})
		}

	case r <= 13:
		for i := range 10 {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: 5, StartHealth: 5,
				reward:          2,
				startDelay:      i * 250,
				speedMultiplier: 1.0,
			})
		}

	case r <= 14:
		for i := range 15 {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: 5, StartHealth: 5,
				reward:          2,
				startDelay:      i * 250,
				speedMultiplier: 1.25,
			})
		}

	case r <= 15:
		for i := range 15 {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: 10, StartHealth: 10,
				reward:          3,
				startDelay:      i * 250,
				speedMultiplier: 1.25,
			})
		}

	case r <= 16:
		for i := range 30 {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: 1, StartHealth: 1,
				reward:          1,
				startDelay:      i * 1000,
				speedMultiplier: 1,
			})
		}

	case r <= 17:
		for i := range 30 {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: 1, StartHealth: 1,
				reward:          1,
				startDelay:      i * 750,
				speedMultiplier: 1.25,
			})
		}

	case r <= 18:
		for i := range 45 {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: 2, StartHealth: 2,
				reward:          2,
				startDelay:      i * 500,
				speedMultiplier: 1.5,
			})
		}

	case r <= 19:
		for i := range 50 {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: 3, StartHealth: 3,
				reward:          2,
				startDelay:      i * 250,
				speedMultiplier: 1.75,
			})
		}

	case r <= 20:
		for i := range 75 {
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: 3, StartHealth: 1,
				reward:          3,
				startDelay:      i * 100,
				speedMultiplier: 2.5,
			})
		}

	default:
		for i := range int(float64(r) * (1 + rand.Float64())) { // r = 10 -> 10 ~ 20 ; r = 100 -> 100 ~ 200
			uid += 1
			game.GS.Enemies = append(game.GS.Enemies, &EnemyObj{
				x: x, y: y, UID: uid, Progress: 0.0,
				Health: max(1, int(float64(r)/5)), StartHealth: max(1, int(float64(r)/5)), // r = 20 -> 4 ; r = 100 -> 20
				reward:          max(1, int(float64(r)/10)), // r = 20 -> 2 ; r = 100 -> 10
				startDelay:      i * max(100, 1100-(r*10)),  // r = 20 -> 900 ; r = 100 -> 100
				speedMultiplier: max(0.1, float64(r)/10),    // r = 20 -> 2.0 ; r = 100 -> 10.0
			})
		}
	}
}
