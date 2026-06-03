package controller

import (
	"math"
	"math/rand"

	"github.com/sverrehu/spacegame/internal/model"
	"github.com/sverrehu/spacegame/internal/utils"
)

const shipRotationSpeed = 2.0            // radians per second
const shipMaxSpeed = 120                 // pixels per second
const shipDeltaSpeed = 22.5              // speed increase per second
const shipPhaserHeatReductionSpeed = 2.5 // phaser heat decrease per second
const shipDamageReductionSpeed = 0.07    // damage decrease per second
const shipInitialBombsLeft = 5
const shipDifferentColors = 20

var shipColors []utils.Color
var nextColorIndex int

func init() {
	shipColors = make([]utils.Color, shipDifferentColors)
	hue := 0.0
	for q := 0; q < shipDifferentColors; q++ {
		shipColors[q] = utils.HSVToColor(hue, 1.0, 1.0)
		hue += 1.0 / shipDifferentColors
	}
	rand.Shuffle(len(shipColors), func(i, j int) {
		shipColors[i], shipColors[j] = shipColors[j], shipColors[i]
	})
}

func (c *Controller) CreateShip(name string) *model.Ship {
	// Called from outside the controller
	c.liveWorld.Mut.Lock()
	defer c.liveWorld.Mut.Unlock()
	loc := c.findGoodLocation(nil)
	ship := model.NewLiveShip(loc, c.nextColor(), 2*math.Pi*rand.Float64(), name, shipInitialBombsLeft)
	c.liveWorld.Ships[ship.Id] = &ship
	return &ship.Ship
}

func (c *Controller) nextColor() utils.Color {
	col := shipColors[nextColorIndex]
	nextColorIndex++
	if nextColorIndex >= len(shipColors) {
		nextColorIndex = 0
	}
	return col
}

func (c *Controller) RemoveShip(shipId int32) {
	// Called from outside the controller
	c.liveWorld.Mut.Lock()
	defer c.liveWorld.Mut.Unlock()
	ship := c.liveWorld.Ships[shipId]
	if ship == nil {
		return
	}
	ship.BaseObject.Delete = true
	ship.BaseObject.Changed = true
	ship.IsAlive = false
}

func (c *Controller) updateShip(ship *model.LiveShip, dt float64) { // dt - delta time (time passed since last update) in seconds
	if !ship.IsAlive {
		return
	}
	c.updateShipDirection(ship, dt)
	c.updateShipAcceleration(ship, dt)
	c.updateShipLocation(ship, dt)
	c.checkCollisionWithBombPacks(ship)
	if ship.PhaserHeat > 0 {
		ship.SubPhaserHeat(shipPhaserHeatReductionSpeed * dt)
	}
	if ship.Damage > 0 {
		ship.SubDamage(shipDamageReductionSpeed * dt)
	}
}

func (c *Controller) updateShipDirection(ship *model.LiveShip, dt float64) {
	if ship.Turn == model.TurnNone {
		return
	}
	dir := ship.Direction
	dir += model.TurnToFactor(ship.Turn) * shipRotationSpeed * dt
	dir = math.Remainder(dir, 2.0*math.Pi)
	if dir < 0.0 {
		dir += 2.0 * math.Pi
	}
	ship.Direction = dir
	ship.BaseObject.Changed = true
}

func (c *Controller) updateShipAcceleration(ship *model.LiveShip, dt float64) {
	if ship.Thrust == model.ThrustNone {
		return
	}
	dir := ship.Direction
	thrustFactor := model.ThrustToFactor(ship.Thrust)
	ship.DriftX += shipDeltaSpeed * thrustFactor * math.Cos(dir)
	ship.DriftY += shipDeltaSpeed * thrustFactor * math.Sin(dir)
	speed := utils.VectorLengthXY(ship.DriftX, ship.DriftY)
	if speed > shipMaxSpeed {
		ship.DriftX *= shipMaxSpeed / speed
		ship.DriftY *= shipMaxSpeed / speed
	}
	ship.BaseObject.Changed = true
}

func (c *Controller) updateShipLocation(ship *model.LiveShip, dt float64) {
	if ship.DriftX == 0.0 && ship.DriftY == 0.0 {
		return
	}
	if ship.Thrust == model.ThrustNone {
		ship.DriftX /= 1.01
		ship.DriftY /= 1.01
		if math.Abs(ship.DriftX) < 0.2 {
			ship.DriftX = 0.0
		}
		if math.Abs(ship.DriftY) < 0.2 {
			ship.DriftY = 0.0
		}
	}
	ship.Position.X += ship.DriftX * dt
	ship.Position.Y -= ship.DriftY * dt
	if ship.Position.X < 0.0 {
		ship.Position.X = 0.0
	} else if ship.Position.X >= c.liveWorld.Width {
		ship.Position.X = c.liveWorld.Width - 1.0
	}
	if ship.Position.Y < 0.0 {
		ship.Position.Y = 0.0
	} else if ship.Position.Y >= c.liveWorld.Height {
		ship.Position.Y = c.liveWorld.Height - 1.0
	}
	ship.BaseObject.Changed = true
}

func (c *Controller) checkCollisionWithBombPacks(ship *model.LiveShip) {
	for _, bombPack := range c.liveWorld.BombPacks {
		if c.isBombPackCollision(ship, bombPack) {
			ship.BombsLeft += bombPack.BombsLeft
			ship.Changed = true
			c.removeBombPack(bombPack)
		}
	}
}

func (c *Controller) isBombPackCollision(ship *model.LiveShip, bombPack *model.LiveBombPack) bool {
	shape := ship.WorldRelativeShape()
	p1 := new(utils.NewPoint(bombPack.Position.X-BombPackRadius, bombPack.Position.Y-BombPackRadius))
	p2 := new(utils.NewPoint(bombPack.Position.X+BombPackRadius, bombPack.Position.Y+BombPackRadius))
	p3 := new(utils.NewPoint(bombPack.Position.X-BombPackRadius, bombPack.Position.Y+BombPackRadius))
	p4 := new(utils.NewPoint(bombPack.Position.X+BombPackRadius, bombPack.Position.Y-BombPackRadius))
	return utils.ShapeAndLineIntersect(shape, p1, p2) || utils.ShapeAndLineIntersect(shape, p3, p4)
}

func (c *Controller) resurrectShip(ship *model.LiveShip) {
	if ship == nil {
		return
	}
	ship.IsAlive = true
	ship.BaseObject.Changed = true
	ship.BombsLeft = shipInitialBombsLeft
	ship.Damage = 0
	ship.PhaserHeat = 0
	ship.DriftX = 0
	ship.DriftY = 0
	ship.Position = c.findGoodLocation(ship)
	ship.Direction = 2 * math.Pi * rand.Float64()
	ship.Turn = model.TurnNone
	ship.Thrust = model.ThrustNone
}

func (c *Controller) registerHit(ship *model.LiveShip, t model.WeaponType, hitter *model.LiveShip, damage float64) {
	if !ship.IsAlive {
		return
	}
	ship.AddDamage(damage)
	if ship.Damage >= 100 {
		ship.AntiScore++
		hitter.Score++
		hitter.Changed = true
		c.killShip(ship)
	}
	c.updatesHandler.HandleHitBy(ship, hitter, t, !ship.IsAlive)
}

func (c *Controller) killShip(ship *model.LiveShip) {
	ship.IsAlive = false
	ship.BaseObject.Changed = true
	c.CreateBombPack(ship)
	c.CreateExplosion(ship)
}

func (c *Controller) findGoodLocation(s *model.LiveShip) utils.Point {
	collidables := c.liveWorld.Collidables(s)
	if len(collidables) == 0 {
		return utils.NewPoint(c.liveWorld.Width/2, c.liveWorld.Height/2)
	}
	minx := c.liveWorld.Width / 100.0
	maxx := c.liveWorld.Width - minx
	miny := c.liveWorld.Height / 100.0
	maxy := c.liveWorld.Height - miny
	x := 0.0
	y := 0.0
	for i := 0; i < 100; i++ {
		x = minx + rand.Float64()*(maxx-minx)
		y = miny + rand.Float64()*(maxy-miny)
		if c.getDistanceToClosestCollidable(collidables, x, y) >= minx*20 {
			break
		}
	}
	return utils.NewPoint(x, y)
}

func (c *Controller) getDistanceToClosestCollidable(collidables []*model.BaseObject, x, y float64) float64 {
	minDist := math.MaxFloat64
	point := new(utils.NewPoint(x, y))
	for _, collidable := range collidables {
		dist := utils.LineLength(point, &collidable.Position)
		if dist < minDist {
			minDist = dist
		}
	}
	return minDist
}

func (c *Controller) isPhaserOverheated(ship *model.LiveShip) bool {
	return ship.PhaserHeat > 75
}
