package server

import (
	"log"
	"net"
	"sync/atomic"

	"github.com/sverrehu/spacegame/internal/controller"
	"github.com/sverrehu/spacegame/internal/network"
)

var lastId atomic.Int32

type ClientAdapter struct {
	transceiver *network.TcpTransceiver
	Id          int32
	ShipId      int32
	ctrl        *controller.Controller
}

func NewClientAdapter(conn net.Conn, ct *controller.Controller) *ClientAdapter {
	a := ClientAdapter{}
	id := lastId.Add(1)
	a.Id = id
	transceiver := network.NewTransceiver(conn, a.HandleIncoming)
	a.transceiver = transceiver
	a.ctrl = ct
	return &a
}

func (a *ClientAdapter) HandleIncoming(msg network.Message) error {
	switch msg.GetType() {
	case network.MessageEnter:
		a.HandleEnterMessage(msg.(network.EnterMessage))
	case network.MessageLeave:
		a.HandleLeaveMessage()
	case network.MessageResurrect:
		a.HandleResurrectMessage()
	case network.MessageTurn:
		a.HandleTurnMessage(msg.(network.TurnMessage))
	case network.MessageThrust:
		a.HandleThrustMessage(msg.(network.ThrustMessage))
	case network.MessageFirePhaser:
		a.HandleFirePhaserMessage()
	case network.MessageFireBomb:
		a.HandleFireBombMessage()
	default:
		panic("Unknown message type")
	}
	return nil
}

func (a *ClientAdapter) HandleEnterMessage(msg network.EnterMessage) {
	log.Printf("%s wants to join the game", msg.Name)
	ship := a.ctrl.CreateShip(msg.Name)
	world, updates := a.ctrl.GetWorldAndUpdates()
	err := a.transceiver.Send(network.NewWelcomeMessage(ship.Id, world.Width, world.Height, updates))
	if err != nil {
		log.Printf("Error sending welcome message: %v", err)
	}
	a.ShipId = ship.Id
	SendInfoToAll(ship.Name + " enters the game")
}

func (a *ClientAdapter) HandleLeaveMessage() {
	if a.ShipId == 0 {
		return
	}
}

func (a *ClientAdapter) HandleResurrectMessage() {
	if a.ShipId == 0 {
		return
	}
	a.ctrl.Resurrect(a.ShipId)
}

func (a *ClientAdapter) HandleThrustMessage(msg network.ThrustMessage) {
	if a.ShipId == 0 {
		return
	}
	a.ctrl.SetThrust(a.ShipId, msg.Thrust)
}

func (a *ClientAdapter) HandleTurnMessage(msg network.TurnMessage) {
	if a.ShipId == 0 {
		return
	}
	a.ctrl.SetTurn(a.ShipId, msg.Turn)
}

func (a *ClientAdapter) HandleFirePhaserMessage() {
	a.ctrl.CreatePhaser(a.ShipId)
}

func (a *ClientAdapter) HandleFireBombMessage() {
	a.ctrl.CreateBomb(a.ShipId)
}

func (a *ClientAdapter) HandleNetworkError(err error) bool {
	if err == nil {
		return false
	}
	log.Printf("Client network error: %v", err)
	name := a.ctrl.GetShipName(a.ShipId)
	a.ctrl.RemoveShip(a.ShipId)
	delete(clientAdapters, a.Id)
	SendInfoToAll(name + " crashed into another dimension")
	a.Close()
	return true
}

func (a *ClientAdapter) Send(msg network.Message) {
	err := a.transceiver.Send(msg)
	if err != nil {
		a.HandleNetworkError(err)
	}
}

func (a *ClientAdapter) Close() {
	a.transceiver.Close()
}
