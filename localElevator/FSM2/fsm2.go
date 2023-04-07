package FSM2

import (
	"Project/config"
	"Project/elevio"
	"Project/localElevator/elevator"
	"time"
)

func FSM2(
	e elevator.Elevator,
	chIoFloor <-chan int,
	chIoObstical <-chan bool,
	chIoStop <-chan bool,
	chUpdatedState chan<- elevator.Elevator,
	chAddButton <-chan elevio.ButtonEvent,
	chRmButton <-chan elevio.ButtonEvent,
) {
	eObj := e
	chUpdatedState <- eObj
	doorTimer := time.NewTimer(0)
	eObj.ClearAllOrders()

	for {
		eObj.UpdateLights()
		select {

		case <-doorTimer.C:
			doorTimer.Reset(config.DoorOpenTime)
		}
	}

}
