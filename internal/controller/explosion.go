package controller

import "github.com/sverrehu/spacegame/internal/model"

const ExplosionMaxRadius = 25
const ExplosionRadiusChangePerSecond = 90

func (c *Controller) CreateExplosion(ship *model.LiveShip) {
	explosion := model.NewLiveExplosion(ship.Position)
	c.liveWorld.Explosions[explosion.Id] = &explosion
}

func (c *Controller) updateExplosion(explosion *model.LiveExplosion, dt float64) { // dt - delta time (time passed since last update) in seconds
	dr := ExplosionRadiusChangePerSecond * dt
	explosion.Changed = true
	if explosion.OuterRadius < ExplosionMaxRadius {
		explosion.OuterRadius += dr
		if explosion.OuterRadius > ExplosionMaxRadius {
			explosion.OuterRadius = ExplosionMaxRadius
		}
	} else {
		if explosion.InnerRadius < ExplosionMaxRadius {
			explosion.InnerRadius += dr
			if explosion.InnerRadius > ExplosionMaxRadius {
				explosion.Delete = true
			}
		}
	}
}
