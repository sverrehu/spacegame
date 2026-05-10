package controller

import (
	"github.com/sverrehu/spacegame/internal/model"
)

const BombPackRadius = 4

func CreateBombPack(ship *model.LiveShip) {
	if ship.BombsLeft == 0 {
		return
	}
	bombPack := model.NewLiveBombPack(ship.Position, ship.Color, ship.BombsLeft)
	liveWorld.BombPacks[bombPack.Id] = &bombPack
}

func updateBombPack(bombPack *model.LiveBombPack, dt float64) { // dt - delta time (time passed since last update) in seconds
}

func removeBombPack(bombPack *model.LiveBombPack) {
	bombPack.BaseObject.Delete = true
	bombPack.BaseObject.Changed = true
}
