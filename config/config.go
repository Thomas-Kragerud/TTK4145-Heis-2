package config

import (
	"time"
)

var NumFloors = 4

const NumButtons = 3

const PollRate = 20 * time.Millisecond

const DoorOpenTime = 3 * time.Second

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}
