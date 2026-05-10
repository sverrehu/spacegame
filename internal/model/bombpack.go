package model

import "github.com/sverrehu/spacegame/internal/utils"

type BombPack struct {
	BaseObject
	BombsLeft int16
}

type LiveBombPack struct {
	BombPack
}

func NewBombPack(position utils.Point, color utils.Color, bombsLeft int16) BombPack {
	return BombPack{BaseObject: NewBaseObject(ObjectBombPack, position, color), BombsLeft: bombsLeft}
}

func NewLiveBombPack(position utils.Point, color utils.Color, bombsLeft int16) LiveBombPack {
	return LiveBombPack{BombPack: NewBombPack(position, color, bombsLeft)}
}

func (o *BombPack) GetType() ObjectType {
	return o.Type
}

func (o *BombPack) ToAnyUpdate(full bool) AnyObjectUpdate {
	return NewAnyObjectUpdate(&o.BaseObject, full)
}

func (o *BombPack) FromAnyUpdate(u AnyObjectUpdate) {
	o.BaseObject.FromAnyUpdate(u)
}
