package server

import (
	"fmt"
	"log"
	"net"
	"strconv"
	"sync"

	"github.com/sverrehu/spacegame/internal/controller"
	"github.com/sverrehu/spacegame/internal/model"
	"github.com/sverrehu/spacegame/internal/network"
)

type updatesHandlerImpl struct {
}

var updatesHandler updatesHandlerImpl
var wg sync.WaitGroup
var clientAdapters map[int32]*ClientAdapter
var ctrl *controller.Controller

func StartServer(port int) {
	updatesHandler = updatesHandlerImpl{}
	clientAdapters = make(map[int32]*ClientAdapter)
	ctrl = controller.NewController()
	ctrl.SetupWorld()
	startListening(port, ctrl)
	wg.Add(1)
	go ctrl.GameLoop(&updatesHandler)
}

func WaitForServer() {
	wg.Wait()
}

func startListening(port int, ctrl *controller.Controller) {
	log.Printf("Starting server on port %d", port)
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal("Error listening:", err)
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Println("Error accepting conn:", err)
				continue
			}
			go handleConnection(conn, ctrl)
		}
	}()
}

func handleConnection(conn net.Conn, ctrl *controller.Controller) {
	log.Printf("Handling connection from %s", conn.RemoteAddr())
	ca := NewClientAdapter(conn, ctrl)
	clientAdapters[ca.Id] = ca
}

func (h *updatesHandlerImpl) HandleUpdates(updates []model.AnyObjectUpdate) {
	if len(updates) == 0 {
		return
	}
	msg := network.NewUpdatesMessage(updates)
	SendToAll(msg)
}

func (h *updatesHandlerImpl) HandleHitBy(ship *model.LiveShip, hitter *model.LiveShip, t model.WeaponType, killed bool) {
	ca := findClientAdapter(ship)
	if ca == nil {
		return
	}
	msg := network.NewHitByMessage(t, killed)
	ca.Send(msg)
	if !killed {
		return
	}
	var text string
	if hitter != nil {
		var weapon string
		if t == model.WeaponPhaser {
			weapon = "phaser"
		} else {
			weapon = "bomb"
		}
		text = fmt.Sprintf("%s was killed by %s's %s.", ship.Name, hitter.Name, weapon)
	} else {
		text = fmt.Sprintf("%s died.", ship.Name)
	}
	SendInfoToAll(text)
}

func SendToAll(msg network.Message) {
	for _, client := range clientAdapters {
		client.Send(msg)
	}
}

func SendInfoToAll(text string) {
	SendToAll(network.NewInGameMessageMessage(network.InGameInfoMessage, text))
}

func findClientAdapter(ship *model.LiveShip) *ClientAdapter {
	for _, client := range clientAdapters {
		if client.ShipId == ship.Id {
			return client
		}
	}
	return nil
}
