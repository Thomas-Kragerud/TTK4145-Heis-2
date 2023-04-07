package messageHandler

import (
	"Project/elevio"
	"Project/localElevator/elevator"
)

type assignTyppe int

const (
	Add    assignTyppe = 0
	Remove assignTyppe = 1
)

type assignValue struct {
	Type     assignTyppe
	BtnEvent elevio.ButtonEvent
}

type networkEvent int

// Gjør til så bostaver når ikke brukes utenfor mappe
const (
	NewCab          networkEvent = 0
	UpdateElevState networkEvent = 1
	NewHall         networkEvent = 2
	AkHall          networkEvent = 3
	ClareHall       networkEvent = 4
	AkRmHall        networkEvent = 5
	PeriodicUpdate  networkEvent = 6
)

type NetworkPackage struct {
	Event    networkEvent
	Elevator elevator.Elevator
	BtnEvent elevio.ButtonEvent
}

type ElevatorUpdate struct {
	Elevator elevator.Elevator
	Alive    bool
	Version  int
}
