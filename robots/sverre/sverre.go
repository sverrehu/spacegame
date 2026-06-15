package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"

	"github.com/sverrehu/spacegame/internal/client"
	"github.com/sverrehu/spacegame/internal/model"
	"github.com/sverrehu/spacegame/internal/utils"
)

const robotName = "I'm Unbeatable!"

const (
	moveNone = iota
	moveForward
	moveBack
)

const (
	stateDead = -1
	stateNone = iota
	stateCollectingBombs
	stateFighting
	stateAvoidingPlace
	stateAvoidingPhaser
	stateAvoidingBomb
)

const (
	absoluteMaxPhaserFires = 1
	absoluteMaxBombFires   = 2
	enemySampleWait        = 2
	enemySamples           = 2
	dropBombsDamageLevel   = 80
	edgeMargin             = 50.0
	logFileName            = "kills.log"
)

type SverresRobot struct {
	client.Robot

	makeLove bool

	currentState int
	heartBeat    float64

	logFile *os.File

	phaserFireDelayCount float64
	phaserFireCount      int
	maxPhaserFires       int
	bombFireDelayCount   float64
	bombFireCount        int
	maxBombFires         int

	currentMove int

	enemy              *model.Ship
	enemyX             []float64
	enemyY             []float64
	enemyMoveDirection float64
	enemySpeed         float64
	enemyLastUpdate    float64
	enemyLocked        bool
	enemyPredictedX    int
	enemyPredictedY    int

	phaser            *model.Phaser
	phaserDirection   float64
	phaserDirectionTo float64

	bombPack    *model.BombPack
	bombPackLoc *utils.Point
}

func (r *SverresRobot) Reset() {
	//TODO implement me
	panic("implement me")
}

func NewSverresRobot() *SverresRobot {
	return &SverresRobot{
		currentState: stateNone,
		currentMove:  moveNone,
		enemyX:       make([]float64, enemySamples),
		enemyY:       make([]float64, enemySamples),
	}
}

func (r *SverresRobot) setMove(move int) {
	if move == r.currentMove {
		return
	}

	switch move {
	case moveNone:
		r.ThrustNone()
	case moveForward:
		r.ThrustForward()
	case moveBack:
		r.ThrustBack()
	}

	r.currentMove = move
}

func (r *SverresRobot) setNewEnemy(ship *model.Ship) {
	r.setMove(moveNone)
	r.enemy = ship

	if ship == nil {
		return
	}

	r.ShowMessage(ship.Name + " is my new enemy!")

	p := r.enemy.Position
	for q := 0; q < enemySamples; q++ {
		r.enemyX[q] = p.X
		r.enemyY[q] = p.Y
	}

	r.enemyMoveDirection = r.enemy.Direction
	r.enemySpeed = 0
	r.enemyLastUpdate = r.heartBeat
	r.setMove(moveForward)
}

func (r *SverresRobot) updateEnemy() {
	currHeartBeat := r.heartBeat
	diffHeartBeat := currHeartBeat - r.enemyLastUpdate

	if diffHeartBeat < enemySampleWait {
		return
	}

	for q := enemySamples - 1; q > 0; q-- {
		r.enemyX[q] = r.enemyX[q-1]
		r.enemyY[q] = r.enemyY[q-1]
	}

	p := r.enemy.Position
	r.enemyX[0] = p.X
	r.enemyY[0] = p.Y

	r.enemyMoveDirection = utils.GetAngle(r.enemyX[1], r.enemyY[1], r.enemyX[0], r.enemyY[0])
	distance := utils.LineLengthXY(r.enemyX[1], r.enemyY[1], r.enemyX[0], r.enemyY[0])
	r.enemySpeed = distance * r.UpdateFrequency() / diffHeartBeat
	r.enemyLastUpdate = currHeartBeat
}

func (r *SverresRobot) chooseEnemy() bool {
	closest := r.ClosestEnemyShip()
	if closest == nil {
		if r.enemy == nil {
			return false
		}
		r.setNewEnemy(nil)
		return true
	}

	if r.enemy == nil || !r.enemy.IsAlive {
		r.setNewEnemy(closest)
		return true
	}

	distCurr := r.MyShip().DistanceTo(&r.enemy.BaseObject)
	distNew := r.MyShip().DistanceTo(&closest.BaseObject)

	if distCurr > 150 && distNew < distCurr-distCurr/2 {
		r.setNewEnemy(closest)
		return true
	}

	return false
}

func (r *SverresRobot) setState(state int) {
	if state != r.currentState {
		switch state {
		case stateDead:
			return
		case stateCollectingBombs:
			r.initCollectingBombs()
		case stateFighting:
			r.initFighting()
		case stateAvoidingPlace:
			r.initAvoidingPlace()
		case stateAvoidingPhaser:
			r.initAvoidingPhaser()
		case stateAvoidingBomb:
			r.initAvoidingBomb()
		}
	}

	r.currentState = state
}

func (r *SverresRobot) handleNone() {
	r.ThrustNone()
	r.TurnNone()
}

func (r *SverresRobot) checkCollectingBombs() bool {
	r.bombPack = r.ClosestBombPack()
	if r.bombPack == nil {
		return false
	}

	enemy := r.ClosestEnemyShip()
	if enemy != nil {
		bpDist := r.MyShip().DistanceTo(&r.bombPack.BaseObject)
		enemyDist := r.MyShip().DistanceTo(&enemy.BaseObject)

		if enemyDist < 180 {
			return false
		}
		if bpDist > enemyDist {
			return false
		}
	}

	r.bombPackLoc = &r.bombPack.Position
	return true
}

func (r *SverresRobot) initCollectingBombs() {}

func (r *SverresRobot) handleCollectingBombs() {
	myDirection := r.MyShip().Direction
	wantedDirection := r.MyShip().AngleTo(&r.bombPack.BaseObject)
	distance := r.MyShip().DistanceTo(&r.bombPack.BaseObject)

	if distance > 10 && distance < 50 &&
		r.GetAbsDeltaAngle(myDirection, wantedDirection) > math.Pi/16.0 {
		r.setMove(moveBack)
	} else {
		r.setMove(moveForward)
	}

	r.turnToward(myDirection, wantedDirection)
}

func (r *SverresRobot) checkFighting() bool {
	r.chooseEnemy()
	return r.enemy != nil
}

func (r *SverresRobot) initFighting() {}

func (r *SverresRobot) handleFighting() {
	myDirection := r.MyShip().Direction
	myLoc := r.MyShip().Position

	enemyDX := float64(r.enemySpeed) * math.Cos(r.enemyMoveDirection)
	enemyDY := float64(r.enemySpeed) * -math.Sin(r.enemyMoveDirection)

	distance := r.MyShip().DistanceTo(&r.enemy.BaseObject)
	time := float64(distance) / MaxPhaserSpeed

	wantedX := r.enemyX[0] + time*enemyDX
	wantedY := r.enemyY[0] + time*enemyDY
	wantedX, wantedY = r.clampToWorld(wantedX, wantedY)

	distance = r.GetDistanceBetween(myLoc.X, myLoc.Y, int(float64(wantedX)+0.5), int(float64(wantedY)+0.5))
	time = float64(distance) / MaxPhaserSpeed

	wantedX = r.enemyX[0] + time*enemyDX
	wantedY = r.enemyY[0] + time*enemyDY
	wantedX, wantedY = r.clampToWorld(wantedX, wantedY)

	r.enemyPredictedX = wantedX
	r.enemyPredictedY = wantedY

	wantedDirection := r.GetAngle(myLoc.X, myLoc.Y, r.enemyPredictedX, r.enemyPredictedY)

	if distance > 150+rand.Float64()*50 {
		r.setMove(moveForward)
	} else {
		r.setMove(moveBack)
	}

	r.turnToward(myDirection, wantedDirection)

	r.enemyLocked = false
	if distance < 400 {
		absDeltaDirection := r.GetAbsDeltaAngle(myDirection, wantedDirection)
		r.enemyLocked = absDeltaDirection < 2.0*math.Pi/32.0
	}
}

func (r *SverresRobot) checkAvoidingPlace() bool {
	if r.currentState == stateCollectingBombs {
		return false
	}

	myLoc := r.MyShip().Position
	margin := edgeMargin

	if r.currentState == stateAvoidingPlace {
		margin += margin
	}

	if myLoc.X < margin {
		return true
	}
	if myLoc.X > r.World().Width-margin-1 {
		return true
	}
	if myLoc.Y < margin {
		return true
	}
	return myLoc.Y > r.World().Height-margin-1
}

func (r *SverresRobot) initAvoidingPlace() {}

func (r *SverresRobot) handleAvoidingPlace() {
	myLoc := r.MyShip().Position
	myDirection := r.MyShip().Direction

	wantedX := r.World().Width / 2
	wantedY := r.World().Height / 2
	wantedDirection := r.GetAngle(myLoc.X, myLoc.Y, wantedX, wantedY)

	r.setMove(moveNone)
	r.setMove(moveForward)
	r.turnToward(myDirection, wantedDirection)
}

func (r *SverresRobot) checkAvoidingPhaser() bool {
	maxDistance := 100 * 100

	var bestPhaser *model.Phaser
	bestDist2 := math.MaxInt
	bestDir := 0.0
	bestDirTo := 0.0

	for _, ph := range r.World().Phasers {
		dist2 := r.GetSquareDistanceBetween(r.MyShip(), ph)
		if dist2 > maxDistance {
			continue
		}

		phDir := ph.Direction
		dirTo := r.GetAngleToObject(ph)

		if r.GetAbsDeltaAngle(dirTo, phDir) < 9.0*math.Pi/10.0 {
			continue
		}

		if dist2 < bestDist2 {
			bestDist2 = dist2
			bestPhaser = ph
			bestDir = phDir
			bestDirTo = dirTo
		}
	}

	if bestPhaser != nil {
		r.phaser = bestPhaser
		r.phaserDirection = bestDir
		r.phaserDirectionTo = bestDirTo
		return true
	}

	return false
}

func (r *SverresRobot) initAvoidingPhaser() {}

func (r *SverresRobot) handleAvoidingPhaser() {
	limit := 9.0 * math.Pi / 10.0
	deltaAngle := r.GetDeltaAngle(r.phaserDirectionTo, r.phaserDirection)

	if r.IsBehind(r.phaser) {
		r.setMove(moveForward)
	} else {
		r.setMove(moveBack)
	}

	if deltaAngle < -limit {
		r.TurnRight()
	} else if deltaAngle > limit {
		r.TurnLeft()
	} else {
		r.TurnNone()
	}
}

func (r *SverresRobot) checkAvoidingBomb() bool {
	bomb := r.ClosestEnemyBomb()
	if bomb == nil {
		return false
	}

	distance := r.GetDistanceBetweenObjects(r.MyShip(), bomb)
	return distance < 150
}

func (r *SverresRobot) initAvoidingBomb() {
	bomb := r.ClosestEnemyBomb()
	if bomb == nil {
		return
	}

	if r.IsBehind(bomb) {
		r.setMove(moveForward)
		if r.IsToTheLeft(bomb) {
			r.TurnRight()
		} else {
			r.TurnLeft()
		}
	} else {
		r.setMove(moveBack)
		if r.IsToTheLeft(bomb) {
			r.TurnLeft()
		} else {
			r.TurnRight()
		}
	}
}

func (r *SverresRobot) handleAvoidingBomb() {}

func (r *SverresRobot) possiblyChangeState() {
	if r.checkAvoidingBomb() {
		r.setState(stateAvoidingBomb)
	} else if r.checkAvoidingPhaser() {
		r.setState(stateAvoidingPhaser)
	} else if r.checkAvoidingPlace() {
		r.setState(stateAvoidingPlace)
	} else if r.checkCollectingBombs() {
		r.setState(stateCollectingBombs)
	} else if r.checkFighting() {
		r.setState(stateFighting)
	} else {
		r.setState(stateNone)
	}
}

func (r *SverresRobot) myFirePhaser() {
	if r.makeLove {
		return
	}
	if r.phaserFireDelayCount > 0 {
		return
	}

	if r.phaserFireCount < r.maxPhaserFires {
		r.phaserFireCount++
		r.FirePhaser()
	} else {
		r.maxPhaserFires = 3 + int(rand.Float64()*float64(absoluteMaxPhaserFires-3))
		r.phaserFireDelayCount = r.UpdateFrequency() / (7 + rand.Float64()*35)
		r.phaserFireCount = 0
	}
}

func (r *SverresRobot) myFireBomb() {
	if r.makeLove {
		return
	}
	if r.bombFireDelayCount > 0 {
		return
	}

	if r.bombFireCount < r.maxBombFires {
		r.bombFireCount++
		r.FireBomb()
	} else {
		r.maxBombFires = 1 + int(rand.Float64()*float64(absoluteMaxBombFires-1))
		r.bombFireDelayCount = r.UpdateFrequency() * (1 + rand.Float64()*2)
		r.bombFireCount = 0
	}
}

func (r *SverresRobot) possiblyShootSomeone() {
	if r.enemyLocked {
		r.myFirePhaser()
	}

	myLoc := r.MyShip().Position
	myDirection := r.MyShip().Direction
	players := r.GetPlayers()

	for q := len(players) - 1; q >= 0; q-- {
		p := players[q]
		if p == r.MyShip() || !p.IsAlive {
			continue
		}

		enemyLoc := p.Position
		distance := r.GetDistanceBetween(myLoc.X, myLoc.Y, enemyLoc.X, enemyLoc.Y)
		if distance > 200 {
			continue
		}

		directionToEnemy := r.GetAngle(myLoc.X, myLoc.Y, enemyLoc.X, enemyLoc.Y)
		absDeltaDirection := r.GetAbsDeltaAngle(myDirection, directionToEnemy)

		if (p != r.enemy || r.currentState == stateAvoidingBomb) &&
			absDeltaDirection < math.Pi/6.0 {
			r.myFirePhaser()
		}

		if absDeltaDirection < math.Pi/4.0 {
			r.myFireBomb()
		}
	}
}

func (r *SverresRobot) Update() {
	r.heartBeat++

	if r.phaserFireDelayCount > 0 {
		r.phaserFireDelayCount--
	}
	if r.bombFireDelayCount > 0 {
		r.bombFireDelayCount--
	}

	r.enemyLocked = false

	r.possiblyChangeState()

	switch r.currentState {
	case stateDead:
		return
	case stateNone:
		r.handleNone()
	case stateCollectingBombs:
		r.handleCollectingBombs()
	case stateFighting:
		r.handleFighting()
	case stateAvoidingPlace:
		r.handleAvoidingPlace()
	case stateAvoidingPhaser:
		r.handleAvoidingPhaser()
	case stateAvoidingBomb:
		r.handleAvoidingBomb()
	}

	r.possiblyShootSomeone()

	if (r.currentState == stateFighting ||
		r.currentState == stateAvoidingPhaser ||
		r.currentState == stateAvoidingBomb) &&
		r.MyShip().Damage > dropBombsDamageLevel {
		r.FireBomb()
	}
}

func (r *SverresRobot) InfoLoggedIntoServer() {
	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to open %s: %v\n", logFileName, err)
		return
	}

	r.logFile = file
}

func (r *SverresRobot) InfoPlayerMoved(player *Player) {
	if player == r.enemy {
		r.updateEnemy()
	}
}

func (r *SverresRobot) InfoPlayerDied(player, killer *Player, weapon byte) {
	if player == r.MyShip() {
		r.phaserFireDelayCount = 0
		r.bombFireDelayCount = 0
		r.setNewEnemy(nil)
	}

	if r.logFile != nil {
		playerName := "?"
		killerName := "?"

		if player != nil {
			playerName = player.GetName()
		}
		if killer != nil {
			killerName = killer.GetName()
		}

		weaponName := WeaponTypeName(weapon)
		_, _ = fmt.Fprintf(r.logFile, "%s\t%s\t%s\r\n", playerName, killerName, weaponName)
	}
}

func (r *SverresRobot) turnToward(myDirection, wantedDirection float64) {
	switch r.GetBestTurn(myDirection, wantedDirection) {
	case TurnLeft:
		r.TurnLeft()
	case TurnRight:
		r.TurnRight()
	default:
		r.TurnNone()
	}
}

func (r *SverresRobot) clampToWorld(x, y float64) (float64, float64) {
	world := r.World()
	if x < 0 {
		x = 0
	} else if x >= world.Width {
		x = world.Width - 1
	}

	if y < 0 {
		y = 0
	} else if y >= world.Height {
		y = world.Height - 1
	}

	return x, y
}

func main() {
	bot := NewSverresRobot()
	bot.Start()
}
