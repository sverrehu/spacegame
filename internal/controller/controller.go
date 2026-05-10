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

var liveWorld model.LiveWorld
var updatesHandler UpdatesHandler

func SetupTestWorld() {
	SetupWorld()
}

func SetupWorld() {
	liveWorld = model.NewLiveWorld(DefaultWidth, DefaultHeight)
	createStars()
}

func GameLoop(f UpdatesHandler) {
	log.Println("Game loop started")
	updatesHandler = f
	msPerUpdate := int64(1000) / UpdatesPerSecond
	for {
		msStart := time.Now().UnixMilli()
		liveWorld.Mut.Lock()
		nextIteration()
		updates := liveWorld.ToAnyUpdates(false)
		liveWorld.PrepareNextRound()
		liveWorld.Mut.Unlock()
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

func nextIteration() {
	for _, ship := range liveWorld.Ships {
		updateShip(ship, 1.0/UpdatesPerSecond)
	}
	for _, phaser := range liveWorld.Phasers {
		updatePhaser(phaser, 1.0/UpdatesPerSecond)
	}
	for _, bomb := range liveWorld.Bombs {
		updateBomb(bomb, 1.0/UpdatesPerSecond)
	}
	for _, bombPack := range liveWorld.BombPacks {
		updateBombPack(bombPack, 1.0/UpdatesPerSecond)
	}
	for _, explosion := range liveWorld.Explosions {
		updateExplosion(explosion, 1.0/UpdatesPerSecond)
	}
}

func createStars() {
	numPixels := liveWorld.Width * liveWorld.Height
	numStars := int(numPixels / 35000.0)
	for q := 0; q < numStars; q++ {
		pos := utils.NewPoint(float64(rand.IntN(int(liveWorld.Width))), float64(rand.IntN(int(liveWorld.Height))))
		hue := rand.Float64() * 0.1875 /* red to yellow */
		saturation := 0.5 + rand.Float64()*0.5
		value := 0.5 + rand.Float64()*0.5
		color := utils.HSVToColor(hue, saturation, value)
		star := model.NewStar(pos, color)
		liveWorld.Stars[star.Id] = &star
	}
}

func Resurrect(shipId int32) {
	liveWorld.Mut.Lock()
	defer liveWorld.Mut.Unlock()
	resurrectShip(liveWorld.Ships[shipId])
}

func SetTurn(shipId int32, turn model.Turn) {
	liveWorld.Mut.Lock()
	defer liveWorld.Mut.Unlock()
	liveWorld.Ships[shipId].Turn = turn
}

func SetThrust(shipId int32, thrust model.Thrust) {
	liveWorld.Mut.Lock()
	defer liveWorld.Mut.Unlock()
	liveWorld.Ships[shipId].Thrust = thrust
}

func GetWorldAndUpdates() (model.World, []model.AnyObjectUpdate) {
	liveWorld.Mut.Lock()
	defer liveWorld.Mut.Unlock()
	world := liveWorld.ToWorld()
	updates := liveWorld.ToAnyUpdates(true)
	return world, updates
}

func GetShipName(id int32) string {
	liveWorld.Mut.Lock()
	defer liveWorld.Mut.Unlock()
	ship := liveWorld.Ships[id]
	if ship == nil {
		return "Someone"
	}
	return ship.Name
}

func findCollidingShip(exceptShipId int32, oldPos, newPos *utils.Point) *model.LiveShip {
	for _, ship := range liveWorld.Ships {
		if ship.Id == exceptShipId {
			continue
		}
		if !ship.IsAlive {
			continue
		}
		if utils.ShapeAndLineIntersect(ship.GetWorldRelativeShape(), oldPos, newPos) {
			return ship
		}
	}
	return nil
}
