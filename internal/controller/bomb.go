package controller

import (
	"math"
	"math/rand/v2"
	"time"

	"github.com/sverrehu/spacegame/internal/model"
	"github.com/sverrehu/spacegame/internal/utils"
)

const bombRotationSpeed = 7.87 // radians per second
const bombDeltaSpeed = 60
const bombMaxSpeed = 150    // pixels per second
const bombMaxDistance = 750 // pixels
const bombShipOffset = 13
const bombFlipSpeed = 7 // flips per second

func CreateBomb(shipId int32) *model.Bomb {
	// Called from outside the controller
	liveWorld.Mut.Lock()
	defer liveWorld.Mut.Unlock()
	owner := liveWorld.Ships[shipId]
	if owner == nil {
		return nil
	}
	if owner.BombsLeft == 0 {
		return nil
	}
	dir := owner.Direction
	x := owner.Position.X + bombShipOffset*math.Cos(dir) + 0.5
	y := owner.Position.Y - bombShipOffset*math.Sin(dir) + 0.5
	bomb := model.NewLiveBomb(utils.NewPoint(x, y), owner.Color)
	bomb.ShipId = owner.Id
	bomb.Direction = owner.Direction
	bomb.Dx = bombMaxSpeed * math.Cos(dir)
	bomb.Dy = bombMaxSpeed * math.Sin(dir)
	liveWorld.Bombs[bomb.Id] = &bomb
	owner.BombsLeft--
	owner.Changed = true
	return &bomb.Bomb
}

func updateBomb(bomb *model.LiveBomb, dt float64) { // dt - delta time (time passed since last update) in seconds
	updateBombFlip(bomb)
	updateBombDirection(bomb, dt)
	updateBombLocation(bomb, dt)
}

func updateBombFlip(bomb *model.LiveBomb) {
	mod := int(float64(time.Now().UnixMilli())/(1.0/bombFlipSpeed)/1000.0) % 2
	wantedFlip := false
	if mod == 0 {
		wantedFlip = true
	}
	if wantedFlip != bomb.Flip {
		bomb.Flip = wantedFlip
		bomb.Changed = true
	}
}

func updateBombDirection(bomb *model.LiveBomb, dt float64) {
	owner := liveWorld.Ships[bomb.ShipId]
	enemyShip := liveWorld.GetClosestShip(owner, &bomb.Position)
	if enemyShip == nil {
		return
	}
	// get direction to the enemy.
	wantedDirection := utils.GetAngle(bomb.Position.X, bomb.Position.Y, enemyShip.Position.X, enemyShip.Position.Y)
	/* determine turn */
	dir := bomb.Direction
	ddir := wantedDirection - bomb.Direction
	maxTurn := bombRotationSpeed * dt
	addir := math.Min(math.Abs(ddir), maxTurn)
	if addir < math.Pi {
		if ddir < 0.0 {
			dir -= addir
		} else {
			dir += addir
		}
	} else {
		if ddir < 0.0 {
			dir += addir
		} else {
			dir -= addir
		}
	}
	dir = math.Remainder(dir, 2.0*math.Pi)
	if dir < 0.0 {
		dir += 2.0 * math.Pi
	}
	bomb.Direction = dir
	bomb.Dx += bombDeltaSpeed * math.Cos(dir)
	bomb.Dy += bombDeltaSpeed * math.Sin(dir)
	speed := utils.VectorLengthXY(bomb.Dx, bomb.Dy)
	if speed > bombMaxSpeed {
		bomb.Dx *= bombMaxSpeed / speed
		bomb.Dy *= bombMaxSpeed / speed
	}
}

func updateBombLocation(bomb *model.LiveBomb, dt float64) {
	oldPos := bomb.Position
	bomb.Position.X += bomb.Dx * dt
	bomb.Position.Y -= bomb.Dy * dt
	victim := findCollidingShip(bomb.ShipId, &oldPos, &bomb.Position)
	if victim != nil {
		bomb.Delete = true
		registerHit(victim, model.WeaponBomb, liveWorld.Ships[bomb.ShipId], 30+rand.Float64()*30.0)
	}
	bomb.Distance += utils.LineLength(&oldPos, &bomb.Position)
	if bomb.Distance >= bombMaxDistance || liveWorld.IsOutside(&bomb.Position) {
		bomb.Delete = true
	}
	bomb.Changed = true
}
