package model

import (
	"math"
	"sync"
	"sync/atomic"

	"github.com/sverrehu/spacegame/internal/utils"
)

var lastId atomic.Int32

type World struct {
	Width, Height float64
	Stars         map[int32]*Star
	Ships         map[int32]*Ship
	Phasers       map[int32]*Phaser
	Bombs         map[int32]*Bomb
	BombPacks     map[int32]*BombPack
	Explosions    map[int32]*Explosion
}

type LiveWorld struct {
	Width, Height float64
	Stars         map[int32]*Star
	Ships         map[int32]*LiveShip
	Phasers       map[int32]*LivePhaser
	Bombs         map[int32]*LiveBomb
	BombPacks     map[int32]*LiveBombPack
	Explosions    map[int32]*LiveExplosion
	Mut           sync.Mutex
}

func NewWorld(width, height float64) World {
	return World{
		Width:      width,
		Height:     height,
		Stars:      make(map[int32]*Star),
		Ships:      make(map[int32]*Ship),
		Phasers:    make(map[int32]*Phaser),
		Bombs:      make(map[int32]*Bomb),
		BombPacks:  make(map[int32]*BombPack),
		Explosions: make(map[int32]*Explosion),
	}
}

func NewLiveWorld(width, height float64) LiveWorld {
	return LiveWorld{
		Width:      width,
		Height:     height,
		Stars:      make(map[int32]*Star),
		Ships:      make(map[int32]*LiveShip),
		Phasers:    make(map[int32]*LivePhaser),
		Bombs:      make(map[int32]*LiveBomb),
		BombPacks:  make(map[int32]*LiveBombPack),
		Explosions: make(map[int32]*LiveExplosion),
	}
}

func (w *LiveWorld) IsOutside(pos *utils.Point) bool {
	return pos.X < 0 || pos.Y < 0 || pos.X >= w.Width || pos.Y >= w.Height
}

func (w *LiveWorld) ToWorld() World {
	world := NewWorld(w.Width, w.Height)
	for _, obj := range w.Stars {
		world.Stars[obj.Id] = obj
	}
	for _, obj := range w.Ships {
		world.Ships[obj.Id] = &obj.Ship
	}
	for _, obj := range w.Phasers {
		world.Phasers[obj.Id] = &obj.Phaser
	}
	for _, obj := range w.Bombs {
		world.Bombs[obj.Id] = &obj.Bomb
	}
	for _, obj := range w.BombPacks {
		world.BombPacks[obj.Id] = &obj.BombPack
	}
	for _, obj := range w.Explosions {
		world.Explosions[obj.Id] = &obj.Explosion
	}
	return world
}

func (w *LiveWorld) PrepareNextRound() {
	// Delete deleted objects, clear changed status.
	for _, obj := range w.Stars {
		obj.Changed = false
		obj.New = false
		if obj.Delete {
			delete(w.Stars, obj.Id)
		}
	}
	for _, obj := range w.Ships {
		obj.Changed = false
		obj.New = false
		if obj.Delete {
			delete(w.Ships, obj.Id)
		}
	}
	for _, obj := range w.Phasers {
		obj.Changed = false
		obj.New = false
		if obj.Delete {
			delete(w.Phasers, obj.Id)
		}
	}
	for _, obj := range w.Bombs {
		obj.Changed = false
		obj.New = false
		if obj.Delete {
			delete(w.Bombs, obj.Id)
		}
	}
	for _, obj := range w.BombPacks {
		obj.Changed = false
		obj.New = false
		if obj.Delete {
			delete(w.BombPacks, obj.Id)
		}
	}
	for _, obj := range w.Explosions {
		obj.Changed = false
		obj.New = false
		if obj.Delete {
			delete(w.Explosions, obj.Id)
		}
	}
}

func (w *LiveWorld) GetCollidables(except *LiveShip) []*BaseObject {
	objs := make([]*BaseObject, 0, 100)
	for _, obj := range w.Ships {
		if except == nil || obj.Id != except.Id {
			objs = append(objs, &obj.BaseObject)
		}
	}
	for _, obj := range w.Phasers {
		objs = append(objs, &obj.BaseObject)
	}
	for _, obj := range w.Bombs {
		objs = append(objs, &obj.BaseObject)
	}
	for _, obj := range w.BombPacks {
		objs = append(objs, &obj.BaseObject)
	}
	return objs
}

func (w *LiveWorld) GetClosestShip(except *LiveShip, pos *utils.Point) *LiveShip {
	shortestDistance := math.MaxFloat64
	var closest *LiveShip = nil
	for _, ship := range w.Ships {
		if except != nil && ship.Id == except.Id {
			continue
		}
		if !ship.IsAlive {
			continue
		}
		distance := utils.LineLength(pos, &ship.Position)
		if distance < shortestDistance {
			shortestDistance = distance
			closest = ship
		}
	}
	return closest
}

func (w *LiveWorld) ToAnyUpdates(full bool) []AnyObjectUpdate {
	updates := make([]AnyObjectUpdate, 0, 100)
	for _, obj := range w.ToGameObjects(full) {
		updates = append(updates, obj.ToAnyUpdate(full || obj.IsChanged()))
	}
	return updates
}

func (w *LiveWorld) ToGameObjects(full bool) []GameObject {
	objs := make([]GameObject, 0, 100)
	if full {
		for _, obj := range w.Stars {
			objs = append(objs, obj)
		}
	}
	for _, obj := range w.Ships {
		if full || obj.New || obj.Changed {
			objs = append(objs, obj)
		}
	}
	for _, obj := range w.Phasers {
		if full || obj.New || obj.Changed {
			objs = append(objs, obj)
		}
	}
	for _, obj := range w.Bombs {
		if full || obj.New || obj.Changed {
			objs = append(objs, obj)
		}
	}
	for _, obj := range w.BombPacks {
		if full || obj.New || obj.Changed {
			objs = append(objs, obj)
		}
	}
	for _, obj := range w.Explosions {
		if full || obj.New || obj.Changed {
			objs = append(objs, obj)
		}
	}
	return objs
}
