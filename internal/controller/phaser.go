package controller

import (
	"log"
	"math"
	"math/rand/v2"

	"github.com/sverrehu/spacegame/internal/model"
	"github.com/sverrehu/spacegame/internal/utils"
)

const phaserSpeed = 370.0     // pixels per second
const phaserMaxDistance = 700 // pixels
const phaserShipOffset = 13

func CreatePhaser(shipId int32) *model.Phaser {
	// Called from outside the controller
	liveWorld.Mut.Lock()
	defer liveWorld.Mut.Unlock()
	owner := liveWorld.Ships[shipId]
	if owner == nil {
		return nil
	}
	if isPhaserOverheated(owner) {
		return nil
	}
	dir := owner.Direction
	x := owner.Position.X + phaserShipOffset*math.Cos(dir) + 0.5
	y := owner.Position.Y - phaserShipOffset*math.Sin(dir) + 0.5
	phaser := model.NewLivePhaser(utils.NewPoint(x, y))
	phaser.ShipId = owner.Id
	phaser.Dx = phaserSpeed * math.Cos(dir)
	phaser.Dy = phaserSpeed * math.Sin(dir)
	liveWorld.Phasers[phaser.Id] = &phaser
	owner.AddPhaserHeat(1 + rand.Float64()*3)
	if isPhaserOverheated(owner) {
		owner.PhaserHeat = 100
	}
	return &phaser.Phaser
}

func updatePhaser(phaser *model.LivePhaser, dt float64) { // dt - delta time (time passed since last update) in seconds
	oldPos := phaser.Position
	phaser.Position.X += phaser.Dx * dt
	phaser.Position.Y -= phaser.Dy * dt
	phaser.BaseObject.Changed = true
	victim := findCollidingShip(phaser.ShipId, &oldPos, &phaser.Position)
	if victim != nil {
		log.Println("hit")
		phaser.Delete = true
		registerHit(victim, model.WeaponPhaser, liveWorld.Ships[phaser.ShipId], 5+rand.Float64()*15.0)
	}
	phaser.Distance += utils.LineLength(&oldPos, &phaser.Position)
	if phaser.Distance >= phaserMaxDistance || liveWorld.IsOutside(&phaser.Position) {
		phaser.Delete = true
	}
}
