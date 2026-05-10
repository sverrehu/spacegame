package client

import (
	"log"
	"net"
	"strconv"
	"sync"

	"github.com/sverrehu/spacegame/internal/model"
)

var server ServerAdapter
var world model.World
var myShipId int32
var worldReadyMutex sync.Mutex
var worldReadyCond = sync.NewCond(&worldReadyMutex)

func StartClient(host string, port int, name string) {
	server = NewServerAdapter(connect(host, port))
	server.Enter(name)
	startUI()
}

func connect(host string, port int) net.Conn {
	conn, err := net.Dial("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal("Error connecting:", err)
	}
	return conn
}

func welcome(id int32, width, height float64, u []model.AnyObjectUpdate) {
	myShipId = id
	world = model.NewWorld(width, height)
	handleUpdates(u)
	worldReadyMutex.Lock()
	defer worldReadyMutex.Unlock()
	worldReadyCond.Broadcast()
}

func myShip() *model.Ship {
	return world.Ships[myShipId]
}

func waitForWorldReady() {
	worldReadyMutex.Lock()
	defer worldReadyMutex.Unlock()
	worldReadyCond.Wait()
}

func handleUpdates(updates []model.AnyObjectUpdate) {
	for _, u := range updates {
		var obj model.GameObject = nil
		switch u.Type {
		case model.ObjectStar:
			if world.Stars[u.Id] == nil {
				world.Stars[u.Id] = &model.Star{}
			}
			if u.Delete {
				delete(world.Stars, u.Id)
			} else {
				obj = world.Stars[u.Id]
			}
		case model.ObjectShip:
			if world.Ships[u.Id] == nil {
				world.Ships[u.Id] = &model.Ship{}
			}
			if u.Delete {
				delete(world.Ships, u.Id)
			} else {
				obj = world.Ships[u.Id]
			}
		case model.ObjectPhaser:
			if world.Phasers[u.Id] == nil {
				world.Phasers[u.Id] = &model.Phaser{}
			}
			if u.Delete {
				delete(world.Phasers, u.Id)
			} else {
				obj = world.Phasers[u.Id]
			}
		case model.ObjectBomb:
			if world.Bombs[u.Id] == nil {
				world.Bombs[u.Id] = &model.Bomb{}
			}
			if u.Delete {
				delete(world.Bombs, u.Id)
			} else {
				obj = world.Bombs[u.Id]
			}
		case model.ObjectBombPack:
			if world.BombPacks[u.Id] == nil {
				world.BombPacks[u.Id] = &model.BombPack{}
			}
			if u.Delete {
				delete(world.BombPacks, u.Id)
			} else {
				obj = world.BombPacks[u.Id]
			}
		case model.ObjectExplosion:
			if world.Explosions[u.Id] == nil {
				world.Explosions[u.Id] = &model.Explosion{}
			}
			if u.Delete {
				delete(world.Explosions, u.Id)
			} else {
				obj = world.Explosions[u.Id]
			}
		default:
			log.Printf("Unknown update type %d", u.Type)
		}
		if obj != nil {
			obj.FromAnyUpdate(u)
		}
	}
}
