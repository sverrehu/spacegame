package controller

import (
	"log"
	"math/rand/v2"
	"time"

	"github.com/sverrehu/spacegame/internal/model"
	"github.com/sverrehu/spacegame/internal/utils"
)

type UpdatesHandler interface {
	HandleUpdates([]model.AnyObjectUpdate)
	HandleHitBy(ship *model.LiveShip, hitter *model.LiveShip, t model.WeaponType, killed bool)
}

const DefaultWidth = 3000
const DefaultHeight = 1700

const UpdatesPerSecond = 60

type Controller struct {
	liveWorld      model.LiveWorld
	updatesHandler UpdatesHandler
}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) SetupWorld() {
	c.liveWorld = model.NewLiveWorld(DefaultWidth, DefaultHeight)
	c.createStars()
}

func (c *Controller) GameLoop(f UpdatesHandler) {
	log.Println("Game loop started")
	c.updatesHandler = f
	msPerUpdate := int64(1000) / UpdatesPerSecond
	for {
		msStart := time.Now().UnixMilli()
		c.liveWorld.Mut.Lock()
		c.nextIteration()
		updates := c.liveWorld.ToAnyUpdates(false)
		c.liveWorld.PrepareNextRound()
		c.liveWorld.Mut.Unlock()
		f.HandleUpdates(updates)
		msSpent := time.Now().UnixMilli() - msStart
		msSleep := msPerUpdate - msSpent
		if msSleep <= 0 {
			log.Printf("WARNING: Game loop iterations take too long.")
		} else {
			time.Sleep(time.Duration(msSleep) * time.Millisecond)
		}
	}
}

func (c *Controller) nextIteration() {
	for _, ship := range c.liveWorld.Ships {
		c.updateShip(ship, 1.0/UpdatesPerSecond)
	}
	for _, phaser := range c.liveWorld.Phasers {
		c.updatePhaser(phaser, 1.0/UpdatesPerSecond)
	}
	for _, bomb := range c.liveWorld.Bombs {
		c.updateBomb(bomb, 1.0/UpdatesPerSecond)
	}
	for _, bombPack := range c.liveWorld.BombPacks {
		c.updateBombPack(bombPack, 1.0/UpdatesPerSecond)
	}
	for _, explosion := range c.liveWorld.Explosions {
		c.updateExplosion(explosion, 1.0/UpdatesPerSecond)
	}
}

func (c *Controller) createStars() {
	numPixels := c.liveWorld.Width * c.liveWorld.Height
	numStars := int(numPixels / 35000.0)
	for q := 0; q < numStars; q++ {
		pos := utils.NewPoint(float64(rand.IntN(int(c.liveWorld.Width))), float64(rand.IntN(int(c.liveWorld.Height))))
		hue := rand.Float64() * 0.1875 /* red to yellow */
		saturation := 0.5 + rand.Float64()*0.5
		value := 0.5 + rand.Float64()*0.5
		color := utils.HSVToColor(hue, saturation, value)
		star := model.NewStar(pos, color)
		c.liveWorld.Stars[star.Id] = &star
	}
}

func (c *Controller) Resurrect(shipId int32) {
	c.liveWorld.Mut.Lock()
	defer c.liveWorld.Mut.Unlock()
	c.resurrectShip(c.liveWorld.Ships[shipId])
}

func (c *Controller) SetTurn(shipId int32, turn model.Turn) {
	c.liveWorld.Mut.Lock()
	defer c.liveWorld.Mut.Unlock()
	c.liveWorld.Ships[shipId].Turn = turn
}

func (c *Controller) SetThrust(shipId int32, thrust model.Thrust) {
	c.liveWorld.Mut.Lock()
	defer c.liveWorld.Mut.Unlock()
	c.liveWorld.Ships[shipId].Thrust = thrust
}

func (c *Controller) GetWorldAndUpdates() (model.World, []model.AnyObjectUpdate) {
	c.liveWorld.Mut.Lock()
	defer c.liveWorld.Mut.Unlock()
	world := c.liveWorld.ToWorld()
	updates := c.liveWorld.ToAnyUpdates(true)
	return world, updates
}

func (c *Controller) GetShipName(id int32) string {
	c.liveWorld.Mut.Lock()
	defer c.liveWorld.Mut.Unlock()
	ship := c.liveWorld.Ships[id]
	if ship == nil {
		return "Someone"
	}
	return ship.Name
}

func (c *Controller) findCollidingShip(exceptShipId int32, oldPos, newPos *utils.Point) *model.LiveShip {
	for _, ship := range c.liveWorld.Ships {
		if ship.Id == exceptShipId {
			continue
		}
		if !ship.IsAlive {
			continue
		}
		if utils.ShapeAndLineIntersect(ship.WorldRelativeShape(), oldPos, newPos) {
			return ship
		}
	}
	return nil
}
