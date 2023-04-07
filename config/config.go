package config

import "time"

// Change to var when everythig is working
const NumFloors = 4
const NumButtons = 3

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

/* Emil
const NumFloors = 4
const NumButtons = 3
const DoorOpenDuration = 3
const StateUpdatePeriodMs = 500
const ElevatorStuckToleranceSec = 5
const ReconnectTimerSec = 3
const LocalElevator = 0

*/
