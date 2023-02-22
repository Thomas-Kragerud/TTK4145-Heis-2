package elevator

import (
	"Project/localElevator/elevio"
)

type ElevatorBehavior string

const (
	EB_Idle     ElevatorBehavior = "EB_Idle"
	EB_DoorOpen ElevatorBehavior = "EB_DoorOpen"
	EB_Moving   ElevatorBehavior = "EB_Moving"
)

type ClearRequestVariant string

const (
	CV_ALL   ClearRequestVariant = "All"
	CV_InDirn ClearRequestVariant = "InDir"
)

type Config struct {
	ClearRequestVariant ClearRequestVariant
	DoorOpenDuration_s    uint32
}

type Elevator struct {
	Floor int
	Dirn elevio.MotorDirection
	Requests [][] bool
	Behavior ElevatorBehavior
	Config   Config
}

func Elevator_uninitialized() Elevator {
	config := Config{ClearRequestVariant: CV_ALL, DoorOpenDuration_s: 3.0}
	requests := make([][]bool, 0)
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		requests = append(requests, make([]bool, elevio.N_BUTTONS))
		for button := range requests[floor] {
			requests[floor][button] = false
		}
	}
	elevator := Elevator{Floor: -1,
	Dirn: elevio.MD_Stop,
	Requests: requests,
	Behavior: EB_Idle,
	Config: config}
	return elevator
}




