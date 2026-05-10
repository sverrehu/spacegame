package client

import (
	"slices"
	"time"
)

type GameMessage struct {
	Text       string
	ExpireTime time.Time
}

type GameMessages struct {
	Messages []*GameMessage
}

func (m *GameMessage) IsExpired() bool {
	return time.Now().After(m.ExpireTime)
}

func NewGameMessages() *GameMessages {
	return &GameMessages{}
}

func (m *GameMessages) Add(text string, duration time.Duration) {
	m.Messages = append(m.Messages, new(GameMessage{text, time.Now().Add(duration)}))
}

func (m *GameMessages) GetMessages() []GameMessage {
	m.Expire()
	messages := make([]GameMessage, len(m.Messages))
	for i, msg := range m.Messages {
		messages[i] = *msg
	}
	return messages
}

func (m *GameMessages) Expire() {
	m.Messages = slices.DeleteFunc(m.Messages, func(msg *GameMessage) bool {
		return msg.IsExpired()
	})
}

func (m *GameMessages) Clear() {
	m.Messages = make([]*GameMessage, 0)
}
