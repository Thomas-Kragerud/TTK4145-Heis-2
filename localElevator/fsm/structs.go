package fsm

import (
	"Project/elevio"
	"Project/localElevator/elevator"
)

type FsmEvent int

const (
	ClearBtn    FsmEvent = 0
	ClearHall   FsmEvent = 1
	ClearCab    FsmEvent = 2
	NewState    FsmEvent = 3
	Update      FsmEvent = 4
	Obstruction FsmEvent = 5
)

type FsmOutput struct {
	Elevator elevator.Elevator
	Event    FsmEvent
	BtnEvent elevio.ButtonEvent
}
