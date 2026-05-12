package network

import (
	"bytes"
	"log"

	"github.com/sverrehu/spacegame/internal/model"
)

type InGameMessageType int

const (
	InGameInfoMessage InGameMessageType = iota
	InGameChatMessage
)

type MessageType int

const (
	// server-to-client messages
	MessageWelcome MessageType = iota
	MessageUpdates
	MessageHitBy
	MessageInGameMessage

	// client-to-server messages
	MessageEnter
	MessageLeave
	MessageResurrect
	MessageTurn
	MessageThrust
	MessageFirePhaser
	MessageFireBomb

	// internal messages
	MessageAbort
)

type Message interface {
	GetType() MessageType
}

type BaseMessage struct {
	Type MessageType
}

type WelcomeMessage struct {
	BaseMessage
	Id      int32
	Width   float64
	Height  float64
	Updates []model.AnyObjectUpdate
}

type UpdatesMessage struct {
	BaseMessage
	Updates []model.AnyObjectUpdate
}

type HitByMessage struct {
	BaseMessage
	Type   model.WeaponType
	Killed bool
}

type InGameMessageMessage struct {
	BaseMessage
	Type InGameMessageType
	Text string
}

type EnterMessage struct {
	BaseMessage
	Name string
}

type LeaveMessage struct {
	BaseMessage
}

type ResurrectMessage struct {
	BaseMessage
}

type TurnMessage struct {
	BaseMessage
	Turn model.Turn
}

type ThrustMessage struct {
	BaseMessage
	Thrust model.Thrust
}

type FirePhaserMessage struct {
	BaseMessage
}

type FireBombMessage struct {
	BaseMessage
}

func (m BaseMessage) GetType() MessageType {
	return m.Type
}

func NewMessage(mt MessageType) BaseMessage {
	return BaseMessage{mt}
}

func NewWelcomeMessage(id int32, width, height float64, updates []model.AnyObjectUpdate) WelcomeMessage {
	return WelcomeMessage{BaseMessage: NewMessage(MessageWelcome), Id: id, Width: width, Height: height, Updates: updates}
}

func NewUpdatesMessage(updates []model.AnyObjectUpdate) UpdatesMessage {
	return UpdatesMessage{BaseMessage: NewMessage(MessageUpdates), Updates: updates}
}

func NewHitByMessage(t model.WeaponType, killed bool) HitByMessage {
	return HitByMessage{BaseMessage: NewMessage(MessageHitBy), Type: t, Killed: killed}
}

func NewInGameMessageMessage(t InGameMessageType, text string) InGameMessageMessage {
	return InGameMessageMessage{BaseMessage: NewMessage(MessageInGameMessage), Type: t, Text: text}
}

func NewEnterMessage(name string) EnterMessage {
	return EnterMessage{BaseMessage: NewMessage(MessageEnter), Name: name}
}

func NewResurrectMessage() ResurrectMessage {
	return ResurrectMessage{BaseMessage: NewMessage(MessageResurrect)}
}

func NewTurnMessage(turn model.Turn) TurnMessage {
	return TurnMessage{BaseMessage: NewMessage(MessageTurn), Turn: turn}
}

func NewThrustMessage(thrust model.Thrust) ThrustMessage {
	return ThrustMessage{BaseMessage: NewMessage(MessageThrust), Thrust: thrust}
}

func NewFirePhaserMessage() FirePhaserMessage {
	return FirePhaserMessage{BaseMessage: NewMessage(MessageFirePhaser)}
}

func NewFireBombMessage() FireBombMessage {
	return FireBombMessage{BaseMessage: NewMessage(MessageFireBomb)}
}

func EncodeMessage(msg Message) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(byte(msg.GetType()))
	switch msg.GetType() {
	case MessageWelcome:
		err := encodeInt32(buf, msg.(WelcomeMessage).Id)
		if err != nil {
			log.Fatal("encode error:", err)
		}
		err = encodeFloat64(buf, msg.(WelcomeMessage).Width)
		if err != nil {
			log.Fatal("encode error:", err)
		}
		err = encodeFloat64(buf, msg.(WelcomeMessage).Height)
		if err != nil {
			log.Fatal("encode error:", err)
		}
		err = encodeUpdates(buf, msg.(WelcomeMessage).Updates)
		if err != nil {
			log.Fatal("encode error:", err)
		}
	case MessageUpdates:
		err := encodeUpdates(buf, msg.(UpdatesMessage).Updates)
		if err != nil {
			log.Fatal("encode error:", err)
		}
	case MessageHitBy:
		err := encodeHitBy(buf, msg.(HitByMessage).Type, msg.(HitByMessage).Killed)
		if err != nil {
			log.Fatal("encode error:", err)
		}
	case MessageInGameMessage:
		err := encodeInGameMessage(buf, msg.(InGameMessageMessage).Type, msg.(InGameMessageMessage).Text)
		if err != nil {
			log.Fatal("encode error:", err)
		}
	case MessageEnter:
		err := encodeString(buf, msg.(EnterMessage).Name)
		if err != nil {
			log.Fatal("encode error:", err)
		}
	case MessageResurrect:
		// empty message body
	case MessageTurn:
		err := encodeByte(buf, byte(msg.(TurnMessage).Turn))
		if err != nil {
			log.Fatal("encode error:", err)
		}
	case MessageThrust:
		err := encodeByte(buf, byte(msg.(ThrustMessage).Thrust))
		if err != nil {
			log.Fatal("encode error:", err)
		}
	case MessageFirePhaser:
		// empty message body
	case MessageFireBomb:
		// empty message body
	default:
		log.Fatal("unhandled message type:", msg.GetType())
	}
	return buf.Bytes()
}

func DecodeMessage(data []byte) Message {
	buf := bytes.NewBuffer(data)
	mtByte, err := buf.ReadByte()
	if err != nil {
		log.Fatal("decode error:", err)
	}
	messageType := MessageType(mtByte)
	var message Message
	switch messageType {
	case MessageWelcome:
		var id int32
		id, err := decodeInt32(buf)
		if err != nil {
			log.Fatal("decode error:", err)
		}
		var width float64
		var height float64
		width, err = decodeFloat64(buf)
		if err != nil {
			log.Fatal("decode error:", err)
		}
		height, err = decodeFloat64(buf)
		if err != nil {
			log.Fatal("decode error:", err)
		}
		updates, err := decodeUpdates(buf)
		if err != nil {
			log.Fatal("decode error:", err)
		}
		message = NewWelcomeMessage(id, width, height, updates)
	case MessageUpdates:
		updates, err := decodeUpdates(buf)
		if err != nil {
			log.Fatal("decode error:", err)
		}
		message = NewUpdatesMessage(updates)
	case MessageResurrect:
		message = NewResurrectMessage()
	case MessageHitBy:
		t, killed, err := decodeHitBy(buf)
		if err != nil {
			log.Fatal("decode error:", err)
		}
		message = NewHitByMessage(t, killed)
	case MessageInGameMessage:
		t, text, err := decodeInGameMessage(buf)
		if err != nil {
			log.Fatal("decode error:", err)
		}
		message = NewInGameMessageMessage(t, text)
	case MessageEnter:
		name, err := decodeString(buf)
		if err != nil {
			log.Fatal("decode error:", err)
		}
		message = NewEnterMessage(name)
	case MessageTurn:
		b, err := decodeByte(buf)
		if err != nil {
			log.Fatal("decode error:", err)
		}
		message = NewTurnMessage(model.Turn(b))
	case MessageThrust:
		b, err := decodeByte(buf)
		if err != nil {
			log.Fatal("decode error:", err)
		}
		message = NewThrustMessage(model.Thrust(b))
	case MessageFirePhaser:
		message = NewFirePhaserMessage()
	case MessageFireBomb:
		message = NewFireBombMessage()
	default:
		log.Fatal("unhandled message type:", messageType)
	}
	return message
}
