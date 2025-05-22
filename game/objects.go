package game

type (
	GameObj interface {
		// Valid types: `Field`, `Road`, `Tower`, `Enemy`,
		Type() string
		// X, Y
		Cord() (int, int)
		// Color as a string
		Color() color
	}

	FieldObj struct {
		x, y  int
		color color
	}
	RoadObj struct {
		x, y  int
		color color
		// Facing direction
		Direction string
	}
	TowerObj struct {
		x, y  int
		color color
		// Tower name
		Name string
		// Damage multiplier.
		damage int
		// Targeting range in tiles.
		fireRange int
		// Fire when progress hits 1
		fireProgress int
		// Delay fire by this in ms.
		fireDelay int
	}
	EnemyObj struct {
		x, y  int
		color color
		// Progress of the enemy.
		//
		// Every 1 progress represents 1 tile moved.
		Progress float64
		// Delay spawning by this compared to phase start in ms.
		startDelay int
		// Progress 1 every second * this.
		speedMultiplier float64
	}
)

func (obj *FieldObj) Type() string     { return "Field" }
func (obj *FieldObj) Cord() (int, int) { return obj.x, obj.y }
func (obj *FieldObj) Color() color     { return obj.color }

func (obj *RoadObj) Type() string     { return "Road" }
func (obj *RoadObj) Cord() (int, int) { return obj.x, obj.y }
func (obj *RoadObj) Color() color     { return obj.color }

func (obj *TowerObj) Type() string     { return "Tower" }
func (obj *TowerObj) Cord() (int, int) { return obj.x, obj.y }
func (obj *TowerObj) Color() color     { return obj.color }

func (obj *EnemyObj) Type() string     { return "Enemy" }
func (obj *EnemyObj) Cord() (int, int) { return obj.x, obj.y }
func (obj *EnemyObj) Color() color     { return obj.color }
