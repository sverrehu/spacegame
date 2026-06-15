package client

import (
	"fmt"
	"math"
	"time"

	"github.com/sverrehu/spacegame/internal/controller"
	"github.com/sverrehu/spacegame/internal/model"
	"github.com/sverrehu/spacegame/internal/utils"
)

type RobotImplementation interface {
	Reset()
	Update()
}

type Robot struct {
	client ClientInterface
	impl   RobotImplementation
}

func (r *Robot) Run(connectString, name string, impl RobotImplementation) {
	r.startClient(connectString, name, false, impl)
}

func (r *Robot) RunWithUI(connectString, name string, impl RobotImplementation) {
	r.startClient(connectString, name, true, impl)
}

func (r *Robot) startClient(connectString, name string, withUI bool, impl RobotImplementation) {
	r.impl = impl
	host, port := utils.HostAndPort(connectString)
	client := NewTcpClient(host, port, name)
	client.Start()
	r.client = client
	go r.updateLoop()
	if withUI {
		// will not return until the UI is closed
		client.StartUI()
	}
}

func (r *Robot) updateLoop() {
	r.impl.Reset()
	msPerUpdate := int64(1000.0 / 10.0)
	resurrectSent := false
	for {
		msStart := time.Now().UnixMilli()
		if !r.MyShip().IsAlive {
			if !resurrectSent {
				r.client.Resurrect()
				resurrectSent = true
			}
		} else {
			resurrectSent = false
			r.impl.Update()
		}
		msSpent := time.Now().UnixMilli() - msStart
		msSleep := msPerUpdate - msSpent
		if msSleep > 0 {
			time.Sleep(time.Duration(msSleep) * time.Millisecond)
		}
	}
}

func (r *Robot) MyShip() *model.Ship {
	return r.client.MyShip()
}

func (r *Robot) World() *model.World {
	return r.client.World()
}

func (r *Robot) TurnLeft() {
	r.client.TurnLeft()
}

func (r *Robot) TurnRight() {
	r.client.TurnRight()
}

func (r *Robot) TurnNone() {
	r.client.TurnNone()
}

func (r *Robot) ThrustForward() {
	r.client.ThrustForward()
}

func (r *Robot) ThrustBack() {
	r.client.ThrustBack()
}

func (r *Robot) ThrustNone() {
	r.client.ThrustNone()
}

func (r *Robot) FirePhaser() {
	r.client.FirePhaser()
}

func (r *Robot) FireBomb() {
	r.client.FireBomb()
}

func (r *Robot) ShowMessage(m string) {
	fmt.Println(m)
}

func (r *Robot) UpdateFrequency() float64 {
	return 1.0 / controller.UpdatesPerSecond
}

func (r *Robot) EnemyShips() []*model.Ship {
	ships := make([]*model.Ship, 0, len(r.client.World().Ships))
	except := r.MyShip()
	for _, ship := range r.client.World().Ships {
		if except != nil && ship.Id == except.Id {
			continue
		}
		if !ship.IsAlive {
			continue
		}
		ships = append(ships, ship)
	}
	return ships
}

func (r *Robot) ClosestEnemyShip() *model.Ship {
	shortestDistance := math.MaxFloat64
	var closest *model.Ship = nil
	for _, ship := range r.EnemyShips() {
		distance := r.MyShip().DistanceTo(&ship.BaseObject)
		if distance < shortestDistance {
			shortestDistance = distance
			closest = ship
		}
	}
	return closest
}

func (r *Robot) ClosestEnemyBomb() *model.Bomb {
	// TODO
	return nil
}

func (r *Robot) ClosestBombPack() *model.BombPack {
	// TODO
	return nil
}

func (r *Robot) closestBaseObject(objects []*model.BaseObject) *model.BaseObject {
	shortestDistance := math.MaxFloat64
	var closest *model.BaseObject = nil
	for _, obj := range objects {
		distance := r.MyShip().DistanceTo(obj)
		if distance < shortestDistance {
			shortestDistance = distance
			closest = obj
		}
	}
	return closest
}
