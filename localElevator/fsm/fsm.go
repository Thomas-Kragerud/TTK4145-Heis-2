package fsm

import (
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
	"Project/localElevator/requests"
)

/*
func setAllLights(e elevator.Elevator) {
	for floor := 0; floor < elevio.N_FLOORS; floor++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			elevio.SetButtonLamp()
		}
	}
}
*/
func FSM(
	ch_orderChan chan elevio.ButtonEvent,
	ch_elevatorState chan<- elevator.Elevator,
	ch_clearLocalHallOrders chan bool,
	ch_arrivedAtFloors chan int, 
	ch_obstruction chan bool,
	ch_timerDoor chan bool) {
	
	elev := elevator.Elevator_uninitialized()

	elevio.SetDoorOpenLamp(false)
	elevio.SetMotorDirection(elevio.MD_Down)

	for {
		floor := <- ch_arrivedAtFloors
		if floor != 0 {
			elevio.SetMotorDirection(elevio.MD_Down)
		} else {
			elevio.SetMotorDirection(elevio.MD_Stop)
			break
		}
	}

	for {
		//Set lights in elevators and such
		select {
		case order := <- ch_orderChan:
			btn_floor := order.Floor
			btn_type := order.Button
			elev = Fsm_onRequestButtonPress(elev,btn_floor,btn_type)
		case floor := <- ch_arrivedAtFloors:
			elev = fsm_onFloorArrival(elev,floor)
		/*
		case <- doorTimer
		*/
		} 
	}

}

func Fsm_onInitBetweenFloors(e elevator.Elevator) {
	elevio.SetMotorDirection(elevio.BT_HallDown)
	e.Dirn = elevio.MD_Down
	e.Behavior = elevator.EB_Moving
}

func Fsm_onRequestButtonPress(e elevator.Elevator, btn_floor int, btn_type elevio.ButtonType) elevator.Elevator{
	switch e.Behavior {
	case elevator.EB_DoorOpen:
		if requests.Requests_shouldClearImmediately(e,btn_floor,btn_type) {
			//TIMER PENIS
		} else {
			e.Requests[btn_floor][btn_type] = true
		}
	case elevator.EB_Moving:
		e.Requests[btn_floor][btn_type] = true
	case elevator.EB_Idle:
		e.Requests[btn_floor][btn_type] = true
		var pair requests.DrinBehaviorPair = requests.Requests_chooseDirection(e)
		e.Dirn = pair.Dirn
		e.Behavior = pair.Behavior
		switch pair.Behavior {
		case elevator.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			//PENIS TIMER
			e = requests.Requests_clearAtCurrentFloor(e)
		case elevator.EB_Moving:
			elevio.SetMotorDirection(e.Dirn)
		}
	}
	//SET ALL LIGHTS 
	return e
}

func fsm_onFloorArrival(e elevator.Elevator, newFloor int) elevator.Elevator{
	e.Floor = newFloor
	elevio.SetFloorIndicator(e.Floor)
	switch e.Behavior {
	case elevator.EB_Moving:
		if requests.Requests_shouldStop(e) {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			e = requests.Requests_clearAtCurrentFloor(e)
			//TIMER FAEN
			//SET ALL LIGHTS AAAAAAAAAAAH
			e.Behavior = elevator.EB_DoorOpen
		}
	}
	return e
}

func fsm_onDoorTimeout(e elevator.Elevator) {
	switch e.Behavior {
	case elevator.EB_DoorOpen:
		pair := requests.Requests_chooseDirection(e)
		e.Dirn = pair.Dirn
		e.Behavior = pair.Behavior

		switch e.Behavior {
		case elevator.EB_DoorOpen:
			//AAAAAAAAAAAAAH TIMER 
			e = requests.Requests_clearAtCurrentFloor(e)
			//SET LIGHS SSSSSSSSS
		case elevator.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(e.Dirn)
		}		
	}
}