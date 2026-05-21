package model

import (
	"github.com/sverrehu/spacegame/internal/utils"
)

type WeaponType int

const (
	WeaponPhaser WeaponType = iota
	WeaponBomb
)

type ObjectType byte

const (
	ObjectStar ObjectType = iota
	ObjectShip
	ObjectPhaser
	ObjectBomb
	ObjectBombPack
	ObjectExplosion
)

type GameObject interface {
	GetType() ObjectType
	IsChanged() bool
	ToAnyUpdate(full bool) AnyObjectUpdate
	FromAnyUpdate(u AnyObjectUpdate)
}

type BaseObject struct {
	Type     ObjectType
	Id       int32
	Position utils.Point
	Color    utils.Color
	New      bool
	Delete   bool
	Changed  bool
}

type AnyObjectUpdate struct {
	Full        bool
	Type        ObjectType
	Id          int32
	Position    utils.Point
	Color       utils.Color
	Direction   float64
	Name        string
	IsAlive     bool
	Score       int16
	AntiScore   int16
	Damage      uint8
	PhaserHeat  uint8
	BombsLeft   int16
	ShipId      int32
	InnerRadius float64
	OuterRadius float64
	Flip        bool
	Delete      bool
}

func NewAnyObjectUpdate(bo *BaseObject, full bool) AnyObjectUpdate {
	return AnyObjectUpdate{
		Full:     full || bo.New,
		Type:     bo.Type,
		Id:       bo.Id,
		Position: bo.Position,
		Color:    bo.Color,
		Delete:   bo.Delete,
	}
}

func NewBaseObject(tp ObjectType, position utils.Point, color utils.Color) BaseObject {
	id := lastId.Add(1)
	return BaseObject{
		Type:     tp,
		Id:       id,
		Position: position,
		Color:    color,
		New:      true,
		Delete:   false,
		Changed:  true,
	}
}

func (o *BaseObject) FromAnyUpdate(u AnyObjectUpdate) {
	o.Type = u.Type
	o.Id = u.Id
	o.Position = u.Position
	if u.Full {
		o.Color = u.Color
	}
	o.Delete = u.Delete
}

func (o *BaseObject) IsChanged() bool {
	return o.New || o.Changed
}

func (o *BaseObject) AngleTo(destination *BaseObject) float64 {
	return utils.GetAngle(o.Position.X, o.Position.Y, destination.Position.X, destination.Position.Y)
}

func (o *BaseObject) DistanceTo(destination *BaseObject) float64 {
	return utils.LineLength(&o.Position, &destination.Position)
}
