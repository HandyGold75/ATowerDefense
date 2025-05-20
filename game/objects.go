package game

type (
	GameObj interface {
		Type() string
		Cord() (int, int)
		Color() color
	}

	FieldObj struct {
		x, y  int
		color color
	}
	RoadObj struct {
		x, y  int
		color color
	}
	TowerObj struct {
		x, y  int
		color color
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
