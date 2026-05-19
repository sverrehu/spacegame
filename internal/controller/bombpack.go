package controller

import (
	"github.com/sverrehu/spacegame/internal/model"
)

const BombPackRadius = 4

func (c *Controller) CreateBombPack(ship *model.LiveShip) {
	if ship.BombsLeft == 0 {
		return
	}
	bombPack := model.NewLiveBombPack(ship.Position, ship.Color, ship.BombsLeft)
	c.liveWorld.BombPacks[bombPack.Id] = &bombPack
}

func (c *Controller) updateBombPack(bombPack *model.LiveBombPack, dt float64) { // dt - delta time (time passed since last update) in seconds
}

func (c *Controller) removeBombPack(bombPack *model.LiveBombPack) {
	bombPack.BaseObject.Delete = true
	bombPack.BaseObject.Changed = true
}
