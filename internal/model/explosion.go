package model

import "github.com/sverrehu/spacegame/internal/utils"

type Explosion struct {
	BaseObject
	InnerRadius float64
	OuterRadius float64
}

type LiveExplosion struct {
	Explosion
}

func NewExplosion(position utils.Point) Explosion {
	return Explosion{BaseObject: NewBaseObject(ObjectExplosion, position, utils.NewColor(1, 1, 0))}
}

func NewLiveExplosion(position utils.Point) LiveExplosion {
	return LiveExplosion{Explosion: NewExplosion(position)}
}

func (o *Explosion) GetType() ObjectType {
	return o.Type
}

func (o *Explosion) ToAnyUpdate(full bool) AnyObjectUpdate {
	u := NewAnyObjectUpdate(&o.BaseObject, full)
	u.InnerRadius = o.InnerRadius
	u.OuterRadius = o.OuterRadius
	return u
}

func (o *Explosion) FromAnyUpdate(u AnyObjectUpdate) {
	o.BaseObject.FromAnyUpdate(u)
	o.InnerRadius = u.InnerRadius
	o.OuterRadius = u.OuterRadius
}
