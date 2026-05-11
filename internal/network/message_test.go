package network

import (
	"testing"

	"github.com/sverrehu/spacegame/internal/model"
	"github.com/sverrehu/spacegame/internal/utils"
)

func TestEncodeDecodeMessage(t *testing.T) {
	msg := NewEnterMessage("foo")
	bytes := EncodeMessage(msg)
	decoded := DecodeMessage(bytes)
	if decoded.GetType() != msg.GetType() {
		t.Errorf("Wrong message type.")
	}
	if decoded.(EnterMessage).Name != msg.Name {
		t.Errorf("Wrong message name, expected: %s, got: %s", msg.Name, decoded.(EnterMessage).Name)
	}
}

func TestEncodeDecodeMessage2(t *testing.T) {
	liveWorld := model.NewLiveWorld(500, 500)
	ship := model.NewLiveShip(utils.NewPoint(0, 0), utils.NewColor(1, 1, 1), 0, "foo", 5)
	liveWorld.Ships[ship.Id] = &ship
	msg := NewWelcomeMessage(30000, liveWorld.Width, liveWorld.Height, liveWorld.ToAnyUpdates(true))
	bytes := EncodeMessage(msg)
	decoded := DecodeMessage(bytes)
	if decoded.GetType() != msg.GetType() {
		t.Errorf("Wrong message type.")
	}
	if decoded.(WelcomeMessage).Width != liveWorld.Width {
		t.Errorf("Wrong width.")
	}
	if len(decoded.(WelcomeMessage).Updates) != 1 {
		t.Errorf("Wrong number of updates.")
	}
}
