package boot

import (
	"Project/elevio"
	"Project/localElevator/elevator"
)

func Elevator(
	pid,
	port string,
	chAtFloor <-chan int,
	numFloors int) elevator.Elevator {
	// Boot elevator
	elevio.Init("localhost:"+port, numFloors)
	eObj := new(elevator.Elevator)
	eObj.Init(pid)
	//chMsgToNetwork <- *eObj

	// Move elevator to closest "certain" floor
	elevio.SetDoorOpenLamp(false)
	elevio.SetMotorDirection(elevio.MD_Down)
	floor := <-chAtFloor
	if floor != 0 {
		for p := floor; p == floor; p = <-chAtFloor {
			continue // continue going down
		}
		eObj.SetFloor(floor - 1)
	} else {
		for p := floor; p == floor; p = <-chAtFloor {
			elevio.SetMotorDirection(elevio.MD_Up)
		}
		eObj.SetFloor(floor + 1)
	}
	elevio.SetMotorDirection(elevio.MD_Stop)
	eObj.SetStateIdle()

	return *eObj
}
