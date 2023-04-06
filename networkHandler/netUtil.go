package networkHandler

import (
	"Project/elevio"
	"Project/localElevator/elevator"
)

type networkEvent int

const (
	NewCab          networkEvent = 0
	UpdateElevState networkEvent = 1
	NewHall         networkEvent = 2
	AkHall          networkEvent = 3
	RmHall          networkEvent = 4
	AkRmHall        networkEvent = 5
	PeriodicUpdate  networkEvent = 6
)

type NetworkPackage struct {
	Event    networkEvent
	Elevator elevator.Elevator
	BtnEvent elevio.ButtonEvent
}
