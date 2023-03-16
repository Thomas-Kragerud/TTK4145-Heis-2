package config

import "Project/elevio"

// Change to var when everythig is working
const NumFloors = 4
const NumButtons = 3

const DoorOpenTime = 3

// InitEnvironment Not implemented
func InitEnvironment() {

}

type SendElev struct {
	Behaviour   string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type InputData struct {
	HallRequests SendHall            `json:"hallRequests"`
	States       map[string]SendElev `json:"states"`
}

type SendHall struct {
	HallRequests [][]bool
}

func (h *SendHall) Init() {
	h.HallRequests = make([][]bool, NumFloors)
	for floor := range h.HallRequests {
		h.HallRequests[floor] = make([]bool, 2)
	}
}

// Update dårlig navn elle??
func (h *SendHall) Update(floor int, button elevio.ButtonType) {
	h.HallRequests[floor][button] = true
}

// Init initialize the elevator
func (e *SendElev) Init() {
	e.CabRequests = make([]bool, NumFloors)
	e.Floor = 0
	e.Behaviour = "idle"
	e.Direction = "stop"
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
