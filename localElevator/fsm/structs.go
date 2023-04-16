package fsm

import (
	"Project/elevio"
	"Project/localElevator/elevator"
)

type FsmEvent int

const (
	ClearHall          FsmEvent = 1
	ClearCab           FsmEvent = 2
	Update             FsmEvent = 4
	Obstruction        FsmEvent = 5
	ClearedObstruction FsmEvent = 6
)

type FsmOutput struct {
	Elevator elevator.Elevator
	Event    FsmEvent
	BtnEvent elevio.ButtonEvent
}
