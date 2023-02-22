package requests

import "Project/localElevator/elevio"
import "Project/localElevator/elevator"

//HEI!

type DrinBehaviorPair struct {
	Dirn     elevio.MotorDirection
	Behavior elevator.ElevatorBehavior
}

func Requests_above(e elevator.Elevator) bool {
	for f := e.Floor + 1; f < elevio.N_FLOORS; f++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if  e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func Requests_below(e elevator.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < elevio.N_BUTTONS; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func Requests_here(e elevator.Elevator) bool {
	for btn := 0; btn < elevio.N_BUTTONS; btn++ {
		if e.Requests[e.Floor][btn] {
			return true
		}
	}
	return false
}

func Requests_chooseDirection(e elevator.Elevator) DrinBehaviorPair {
	switch e.Dirn {
	case elevio.MD_Up:
		if Requests_above(e)  {
			return DrinBehaviorPair{Dirn: elevio.MD_Up, Behavior: elevator.EB_Moving}
		} else if Requests_here(e) {
			return DrinBehaviorPair{Dirn: elevio.MD_Down, Behavior: elevator.EB_DoorOpen}
		} else if Requests_below(e) {
			return DrinBehaviorPair{Dirn: elevio.MD_Down, Behavior: elevator.EB_Moving}
		} else {
			return DrinBehaviorPair{Dirn: elevio.MD_Stop, Behavior: elevator.EB_Idle}
		}
			
	case elevio.MD_Down:
		if Requests_below(e) {
			return DrinBehaviorPair{Dirn: elevio.MD_Down, Behavior: elevator.EB_Moving}
		} else if Requests_here(e) {
			return DrinBehaviorPair{Dirn: elevio.MD_Up, Behavior: elevator.EB_DoorOpen}
		} else if Requests_above(e) {
			return DrinBehaviorPair{Dirn: elevio.MD_Up, Behavior: elevator.EB_Moving}
		} else {
			return DrinBehaviorPair{Dirn: elevio.MD_Stop, Behavior: elevator.EB_Idle}
		}

	case elevio.MD_Stop:
		if Requests_here(e) {
			return DrinBehaviorPair{Dirn: elevio.MD_Stop, Behavior: elevator.EB_DoorOpen}
		} else if Requests_above(e) {
			return DrinBehaviorPair{Dirn: elevio.MD_Up, Behavior: elevator.EB_Moving}
		} else if Requests_below(e) {
			return DrinBehaviorPair{Dirn: elevio.MD_Down, Behavior: elevator.EB_Moving}
		} else {
			return DrinBehaviorPair{Dirn: elevio.MD_Stop, Behavior: elevator.EB_Idle}
		}
	default:
		return DrinBehaviorPair{Dirn: elevio.MD_Stop, Behavior: elevator.EB_Idle}
	}
}

func Requests_shouldStop(e elevator.Elevator) bool {
	switch e.Dirn {
	case elevio.MD_Down:
		return e.Requests[e.Floor][elevio.BT_HallDown] || e.Requests[e.Floor][elevio.BT_Cab] || !Requests_below(e)
	case elevio.MD_Up:
		return e.Requests[e.Floor][elevio.BT_HallUp] || e.Requests[e.Floor][elevio.BT_Cab] || !Requests_above(e)
	default:
		return true
	}
}

func Requests_shouldClearImmediately(e elevator.Elevator, btn_floor int, btn_type elevio.ButtonType) bool{
	switch e.Config.ClearRequestVariant {
	case elevator.CV_ALL:
		return e.Floor == btn_floor
	case elevator.CV_InDirn:
		return e.Floor == btn_floor && ((e.Dirn == elevio.MD_Up && btn_type == elevio.BT_HallUp) || (e.Dirn == elevio.MD_Down && btn_type == elevio.BT_HallDown) || e.Dirn == elevio.MD_Stop || btn_type == elevio.BT_Cab)
	default:
		return false
	}
}

func Requests_clearAtCurrentFloor(e elevator.Elevator) elevator.Elevator {
	switch e.Config.ClearRequestVariant {
	case elevator.CV_ALL:
		for f := 0; f < elevio.N_BUTTONS; f++ {
			e.Requests[e.Floor][f] = false
		}
	case elevator.CV_InDirn:
		e.Requests[e.Floor][elevio.BT_Cab] = false
		switch e.Dirn {
		case elevio.MD_Up:
			if !Requests_above(e) && !e.Requests[e.Floor][elevio.BT_HallUp]{
				e.Requests[e.Floor][elevio.BT_HallDown] = false
			}
			e.Requests[e.Floor][elevio.BT_HallUp] = false
		case elevio.MD_Down:
			if !Requests_below(e) && !e.Requests[e.Floor][elevio.BT_HallDown]{
				e.Requests[e.Floor][elevio.BT_HallUp] = false
			}
			e.Requests[e.Floor][elevio.BT_HallDown] = false
		default:
			e.Requests[e.Floor][elevio.BT_HallUp] = false
			e.Requests[e.Floor][elevio.BT_HallDown] = false
		}
	default:
	}
	return e
}
