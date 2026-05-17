package client

import (
	"log"
	"net"

	"github.com/sverrehu/spacegame/internal/model"
	"github.com/sverrehu/spacegame/internal/network"
)

type IncomingHandler interface {
	welcome(id int32, width, height float64, u []model.AnyObjectUpdate)
	handleUpdates(updates []model.AnyObjectUpdate)
	addInfoMessage(text string)
	addChatMessage(text string)
}

type ServerAdapter struct {
	Transceiver     *network.TcpTransceiver
	incomingHandler IncomingHandler
}

func NewServerAdapter(conn net.Conn, handler IncomingHandler) ServerAdapter {
	a := ServerAdapter{}
	transceiver := network.NewTransceiver(conn, a.HandleIncomingMessage)
	a.Transceiver = transceiver
	a.incomingHandler = handler
	return a
}

func (a *ServerAdapter) Send(msg network.Message) {
	a.handleError(a.Transceiver.Send(msg))
}

func (a *ServerAdapter) handleError(err error) {
	if err != nil {
		log.Fatal("send error:", err)
	}
}

func (a *ServerAdapter) HandleIncomingMessage(msg network.Message) error {
	switch msg.GetType() {
	case network.MessageWelcome:
		a.handleWelcomeMessage(msg.(network.WelcomeMessage))
	case network.MessageUpdates:
		a.handleUpdatesMessage(msg.(network.UpdatesMessage))
	case network.MessageHitBy:
		a.handleHitByMessage(msg.(network.HitByMessage))
	case network.MessageInGameMessage:
		a.handleNewInGameMessageMessage(msg.(network.InGameMessageMessage))
	default:
		panic("Unknown message type")
	}
	return nil
}

func (a *ServerAdapter) handleWelcomeMessage(msg network.WelcomeMessage) {
	log.Printf("Connected with id %d", msg.Id)
	a.incomingHandler.welcome(msg.Id, msg.Width, msg.Height, msg.Updates)
}

func (a *ServerAdapter) handleUpdatesMessage(msg network.UpdatesMessage) {
	a.incomingHandler.handleUpdates(msg.Updates)
}

func (a *ServerAdapter) handleHitByMessage(msg network.HitByMessage) {
	if msg.Killed {
		playBigExplosion()
	} else if msg.Type == model.WeaponBomb {
		playSmallExplosion()
	} else {
		playPhaserHit()
	}
}

func (a *ServerAdapter) handleNewInGameMessageMessage(msg network.InGameMessageMessage) {
	switch msg.Type {
	case network.InGameInfoMessage:
		a.incomingHandler.addInfoMessage(msg.Text)
	case network.InGameChatMessage:
		a.incomingHandler.addChatMessage(msg.Text)
	}
}

func (a *ServerAdapter) sendEnterMessage(name string) {
	log.Println("Connecting to server...")
	m := network.NewEnterMessage(name)
	a.Send(m)
}

func (a *ServerAdapter) sendLeaveMessage() {

}

func (a *ServerAdapter) sendTurnMessage(turn model.Turn) {
	m := network.NewTurnMessage(turn)
	a.Send(m)
}

func (a *ServerAdapter) sendThrustMessage(thrust model.Thrust) {
	m := network.NewThrustMessage(thrust)
	a.Send(m)
}

func (a *ServerAdapter) sendFirePhaserMessage() {
	m := network.NewFirePhaserMessage()
	a.Send(m)
}

func (a *ServerAdapter) sendFireBombMessage() {
	m := network.NewFireBombMessage()
	a.Send(m)
}

func (a *ServerAdapter) sendResurrectMessage() {
	m := network.NewResurrectMessage()
	a.Send(m)
}
