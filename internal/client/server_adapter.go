package client

import (
	"log"
	"net"
	"time"

	"github.com/sverrehu/spacegame/internal/model"
	"github.com/sverrehu/spacegame/internal/network"
)

type ServerAdapter struct {
	Transceiver *network.TcpTransceiver
}

func NewServerAdapter(conn net.Conn) ServerAdapter {
	a := ServerAdapter{}
	transceiver := network.NewTransceiver(conn, a.HandleIncoming)
	a.Transceiver = transceiver
	return a
}

func (a *ServerAdapter) Send(msg network.Message) error {
	return a.Transceiver.Send(msg)
}

func (a *ServerAdapter) HandleIncoming(msg network.Message) error {
	switch msg.GetType() {
	case network.MessageWelcome:
		a.HandleWelcomeMessage(msg.(network.WelcomeMessage))
	case network.MessageUpdates:
		a.HandleUpdatesMessage(msg.(network.UpdatesMessage))
	case network.MessageHitBy:
		a.HandleHitByMessage(msg.(network.HitByMessage))
	case network.MessageInGameMessage:
		a.HandleNewInGameMessageMessage(msg.(network.InGameMessageMessage))
	default:
		panic("Unknown message type")
	}
	return nil
}

func (a *ServerAdapter) HandleWelcomeMessage(msg network.WelcomeMessage) {
	log.Printf("Connected with id %d", msg.Id)
	welcome(msg.Id, msg.Width, msg.Height, msg.Updates)
}

func (a *ServerAdapter) HandleUpdatesMessage(msg network.UpdatesMessage) {
	handleUpdates(msg.Updates)
}

func (a *ServerAdapter) HandleHitByMessage(msg network.HitByMessage) {
	if msg.Killed {
		playBigExplosion()
	} else if msg.Type == model.WeaponBomb {
		playSmallExplosion()
	} else {
		playPhaserHit()
	}
}

func (a *ServerAdapter) handleError(err error) {
	if err != nil {
		log.Fatal("send error:", err)
	}
}

func (a *ServerAdapter) Enter(name string) {
	log.Println("Connecting to server...")
	m := network.NewEnterMessage(name)
	a.handleError(a.Send(m))
}

func (a *ServerAdapter) Leave() {

}

func (a *ServerAdapter) SetTurn(turn model.Turn) {
	m := network.NewTurnMessage(turn)
	a.handleError(a.Send(m))
}

func (a *ServerAdapter) SetThrust(thrust model.Thrust) {
	m := network.NewThrustMessage(thrust)
	a.handleError(a.Send(m))
}

func (a *ServerAdapter) FirePhaser() {
	m := network.NewFirePhaserMessage()
	a.handleError(a.Send(m))
}

func (a *ServerAdapter) FireBomb() {
	m := network.NewFireBombMessage()
	a.handleError(a.Send(m))
}

func (a *ServerAdapter) Resurrect() {
	m := network.NewResurrectMessage()
	a.handleError(a.Send(m))
}

func (a *ServerAdapter) HandleNewInGameMessageMessage(msg network.InGameMessageMessage) {
	switch msg.Type {
	case network.InGameInfoMessage:
		messages.Add(msg.Text, 4*time.Second)
	case network.InGameChatMessage:
		chatMessages.Add(msg.Text, 6*time.Second)
	}
}
