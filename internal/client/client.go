package client

import (
	"log"
	"net"
	"strconv"
	"sync"

	"github.com/sverrehu/spacegame/internal/model"
)

type ClientInterface interface {
	GetWorld() *model.World
	GetMyShip() *model.Ship
	TurnLeft()
	TurnRight()
	TurnNone()
	ThrustForward()
	ThrustBack()
	ThrustNone()
	FirePhaser()
	FireBomb()
	Resurrect()
}

type Client struct {
	host            string
	port            int
	name            string
	server          ServerAdapter
	world           model.World
	myShipId        int32
	worldReadyMutex sync.Mutex
	worldReadyCond  *sync.Cond
}

func NewTcpClient(host string, port int, name string) *Client {
	client := &Client{
		host: host,
		port: port,
		name: name,
	}
	client.worldReadyCond = sync.NewCond(&client.worldReadyMutex)
	return client
}

func (c *Client) Start() {
	c.server = NewServerAdapter(c.connect(c.host, c.port), c)
	c.server.sendEnterMessage(c.name)
	c.waitForWorldReady()
	startUI(c)
}

func (c *Client) connect(host string, port int) net.Conn {
	conn, err := net.Dial("tcp", host+":"+strconv.Itoa(port))
	if err != nil {
		log.Fatal("Error connecting:", err)
	}
	return conn
}

func (c *Client) welcome(id int32, width, height float64, u []model.AnyObjectUpdate) {
	c.myShipId = id
	c.world = model.NewWorld(width, height)
	c.handleUpdates(u)
	c.worldReadyMutex.Lock()
	defer c.worldReadyMutex.Unlock()
	c.worldReadyCond.Broadcast()
}

func (c *Client) waitForWorldReady() {
	c.worldReadyMutex.Lock()
	defer c.worldReadyMutex.Unlock()
	c.worldReadyCond.Wait()
}

func (c *Client) handleUpdates(updates []model.AnyObjectUpdate) {
	for _, u := range updates {
		var obj model.GameObject = nil
		switch u.Type {
		case model.ObjectStar:
			if c.world.Stars[u.Id] == nil {
				c.world.Stars[u.Id] = &model.Star{}
			}
			if u.Delete {
				delete(c.world.Stars, u.Id)
			} else {
				obj = c.world.Stars[u.Id]
			}
		case model.ObjectShip:
			if c.world.Ships[u.Id] == nil {
				c.world.Ships[u.Id] = &model.Ship{}
			}
			if u.Delete {
				delete(c.world.Ships, u.Id)
			} else {
				obj = c.world.Ships[u.Id]
			}
		case model.ObjectPhaser:
			if c.world.Phasers[u.Id] == nil {
				c.world.Phasers[u.Id] = &model.Phaser{}
			}
			if u.Delete {
				delete(c.world.Phasers, u.Id)
			} else {
				obj = c.world.Phasers[u.Id]
			}
		case model.ObjectBomb:
			if c.world.Bombs[u.Id] == nil {
				c.world.Bombs[u.Id] = &model.Bomb{}
			}
			if u.Delete {
				delete(c.world.Bombs, u.Id)
			} else {
				obj = c.world.Bombs[u.Id]
			}
		case model.ObjectBombPack:
			if c.world.BombPacks[u.Id] == nil {
				c.world.BombPacks[u.Id] = &model.BombPack{}
			}
			if u.Delete {
				delete(c.world.BombPacks, u.Id)
			} else {
				obj = c.world.BombPacks[u.Id]
			}
		case model.ObjectExplosion:
			if c.world.Explosions[u.Id] == nil {
				c.world.Explosions[u.Id] = &model.Explosion{}
			}
			if u.Delete {
				delete(c.world.Explosions, u.Id)
			} else {
				obj = c.world.Explosions[u.Id]
			}
		default:
			log.Printf("Unknown update type %d", u.Type)
		}
		if obj != nil {
			obj.FromAnyUpdate(u)
		}
	}
}

func (c *Client) GetWorld() *model.World {
	return &c.world
}

func (c *Client) GetMyShip() *model.Ship {
	return c.world.Ships[c.myShipId]
}

func (c *Client) TurnLeft() {
	c.server.sendTurnMessage(model.TurnLeft)
}

func (c *Client) TurnRight() {
	c.server.sendTurnMessage(model.TurnRight)
}

func (c *Client) TurnNone() {
	c.server.sendTurnMessage(model.TurnNone)
}

func (c *Client) ThrustForward() {
	c.server.sendThrustMessage(model.ThrustForward)
}

func (c *Client) ThrustBack() {
	c.server.sendThrustMessage(model.ThrustBack)
}

func (c *Client) ThrustNone() {
	c.server.sendThrustMessage(model.ThrustNone)
}

func (c *Client) FirePhaser() {
	c.server.sendFirePhaserMessage()
}

func (c *Client) FireBomb() {
	c.server.sendFireBombMessage()
}

func (c *Client) Resurrect() {
	c.server.sendResurrectMessage()
}
