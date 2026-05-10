package model

import "github.com/sverrehu/spacegame/internal/utils"

type Bomb struct {
	BaseObject
	ShipId int32
	Flip   bool
}

type LiveBomb struct {
	Bomb
	Direction float64
	Dx, Dy    float64
	Distance  float64
}

func NewBomb(position utils.Point, color utils.Color) Bomb {
	return Bomb{BaseObject: NewBaseObject(ObjectBomb, position, color)}
}

func NewLiveBomb(position utils.Point, color utils.Color) LiveBomb {
	return LiveBomb{Bomb: NewBomb(position, color)}
}

func (o *Bomb) GetType() ObjectType {
	return o.Type
}

func (o *Bomb) ToAnyUpdate(full bool) AnyObjectUpdate {
	u := NewAnyObjectUpdate(&o.BaseObject, full)
	u.Flip = o.Flip
	return u
}

func (o *Bomb) FromAnyUpdate(u AnyObjectUpdate) {
	o.BaseObject.FromAnyUpdate(u)
	o.Flip = u.Flip
}
