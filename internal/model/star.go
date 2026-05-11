package model

import "github.com/sverrehu/spacegame/internal/utils"

type Star struct {
	BaseObject
}

func NewStar(position utils.Point, color utils.Color) Star {
	return Star{NewBaseObject(ObjectStar, position, color)}
}

func (o *Star) GetType() ObjectType {
	return o.Type
}

func (o *Star) ToAnyUpdate(full bool) AnyObjectUpdate {
	return NewAnyObjectUpdate(&o.BaseObject, full)
}

func (o *Star) FromAnyUpdate(u AnyObjectUpdate) {
	o.BaseObject.FromAnyUpdate(u)
}
