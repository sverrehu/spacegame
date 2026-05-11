package model

import "github.com/sverrehu/spacegame/internal/utils"

type Phaser struct {
	BaseObject
}

type LivePhaser struct {
	Phaser
	Dx, Dy   float64
	Distance float64
	ShipId   int32
}

func NewPhaser(position utils.Point) Phaser {
	return Phaser{BaseObject: NewBaseObject(ObjectPhaser, position, utils.NewColor(1, 1, 1))}
}

func NewLivePhaser(position utils.Point) LivePhaser {
	return LivePhaser{Phaser: NewPhaser(position)}
}

func (o *Phaser) GetType() ObjectType {
	return o.Type
}

func (o *Phaser) ToAnyUpdate(full bool) AnyObjectUpdate {
	return NewAnyObjectUpdate(&o.BaseObject, full)
}

func (o *Phaser) FromAnyUpdate(u AnyObjectUpdate) {
	o.BaseObject.FromAnyUpdate(u)
}
