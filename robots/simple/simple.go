package main

import (
	"fmt"

	"github.com/sverrehu/spacegame/internal/client"
)

type SimpleRobot struct {
	client.Robot
}

func NewSimpleRobot() SimpleRobot {
	return SimpleRobot{}
}

func (r *SimpleRobot) Run() {
	r.Robot.RunWithUI("localhost:9999", "SimpleRobot", r)
}

func (r *SimpleRobot) Reset() {
	// TODO
}

func (r *SimpleRobot) Update() {
	fmt.Println("SimpleRobot.Update")
	r.ThrustForward()
	r.TurnRight()
	r.FirePhaser()
}

func main() {
	new(NewSimpleRobot()).Run()
}
