package model

import "github.com/sverrehu/spacegame/internal/utils"

type Thrust byte

const (
	ThrustNone Thrust = iota
	ThrustForward
	ThrustBack
)

type Turn byte

const (
	TurnNone Turn = iota
	TurnLeft
	TurnRight
)

type Ship struct {
	BaseObject
	Name       string
	Direction  float64
	IsAlive    bool
	Score      int16
	AntiScore  int16
	Damage     float64
	PhaserHeat float64
	BombsLeft  int16
}

type LiveShip struct {
	Ship
	Turn           Turn
	Thrust         Thrust
	DriftX, DriftY float64
}

var ShipShape = []utils.Point{utils.NewPoint(13.0, 0.0), utils.NewPoint(-10.0, 7.0), utils.NewPoint(-10.0, -7.0)}

func NewShip(position utils.Point, color utils.Color, direction float64, name string, bombsLeft int16) Ship {
	return Ship{
		BaseObject: NewBaseObject(ObjectShip, position, color),
		Direction:  direction,
		Name:       name,
		IsAlive:    true,
		BombsLeft:  bombsLeft,
	}
}

func NewLiveShip(position utils.Point, color utils.Color, direction float64, name string, bombsLeft int16) LiveShip {
	return LiveShip{Ship: NewShip(position, color, direction, name, bombsLeft)}
}

func (o *Ship) GetType() ObjectType {
	return o.Type
}

func (o *Ship) ToAnyUpdate(full bool) AnyObjectUpdate {
	aou := NewAnyObjectUpdate(&o.BaseObject, full)
	if full {
		aou.Name = o.Name
	}
	aou.Direction = o.Direction
	aou.IsAlive = o.IsAlive
	aou.Score = o.Score
	aou.AntiScore = o.AntiScore
	aou.Damage = uint8(o.Damage + 0.5)
	aou.PhaserHeat = uint8(o.PhaserHeat + 0.5)
	aou.BombsLeft = o.BombsLeft
	return aou
}

func (o *Ship) FromAnyUpdate(u AnyObjectUpdate) {
	o.BaseObject.FromAnyUpdate(u)
	if u.Full {
		o.Name = u.Name
	}
	o.Direction = u.Direction
	o.IsAlive = u.IsAlive
	o.Score = u.Score
	o.AntiScore = u.AntiScore
	o.Damage = float64(u.Damage)
	o.PhaserHeat = float64(u.PhaserHeat)
	o.BombsLeft = u.BombsLeft
}

func (o *LiveShip) GetType() ObjectType {
	return o.Type
}

func (o *LiveShip) AddPhaserHeat(n float64) {
	o.PhaserHeat += n
	if o.PhaserHeat > 100 {
		o.PhaserHeat = 100
	}
	o.Changed = true
}

func (o *LiveShip) SubPhaserHeat(n float64) {
	o.PhaserHeat -= n
	if o.PhaserHeat < 0 {
		o.PhaserHeat = 0
	}
	o.Changed = true
}

func (o *LiveShip) AddDamage(n float64) {
	o.Damage += n
	if o.Damage > 100 {
		o.Damage = 100
	}
	o.Changed = true
}

func (o *LiveShip) SubDamage(n float64) {
	o.Damage -= n
	if o.Damage < 0 {
		o.Damage = 0
	}
	o.Changed = true
}

func (o *Ship) GetWorldRelativeShape() []utils.Point {
	rotated := utils.Rotate(ShipShape, o.Direction)
	return utils.Tranlate(rotated, &o.Position)
}

func ThrustToFactor(thrust Thrust) float64 {
	switch thrust {
	case ThrustNone:
		return 0
	case ThrustForward:
		return 1
	case ThrustBack:
		return -1
	default:
		panic("invalid thrust")
	}
}

func TurnToFactor(turn Turn) float64 {
	switch turn {
	case TurnNone:
		return 0
	case TurnLeft:
		return 1
	case TurnRight:
		return -1
	default:
		panic("invalid turn")
	}
}
