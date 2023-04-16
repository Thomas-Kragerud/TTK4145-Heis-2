package fsmUtils

import (
	"Project/config"
	"Project/elevio"
	"Project/localElevator/elevator"
	"fmt"
	"log"
)

func GetNextDirection(e *elevator.Elevator) elevio.MotorDirection {
	// Samme retning: Hvis orde forbi posisjon med motsatt retning
	// Samme retning; Hvis cab ordre forbi i samme rettning
	// Idle: Hvis ingen ordre
	// Motsatt retning: Hvis ordre (else)
	if e.OrderIsEmpty() {
		//Kan mekke funksjon som kjører til nærmeste etasje
		return elevio.MD_Stop
	} else {
		switch e.Dir {
		case elevio.MD_Up:
			for f := e.Floor+1; f < config.NumFloors; f++ {
				if e.Orders[f][elevio.BT_HallDown] || e.Orders[f][elevio.BT_Cab] && (e.Floor != config.NumFloors-1) {
					return elevio.MD_Up
				}
			}

		case elevio.MD_Down:
			for f := 0; f < e.Floor; f++ {
				if e.Orders[f][elevio.BT_HallUp] || e.Orders[f][elevio.BT_Cab] && (e.Floor != 0){
					log.Printf("Going down")
					return elevio.MD_Down
				}
			}
		case elevio.MD_Stop:
			if AnyOrderInDirection(e, elevio.MD_Down) {
				return elevio.MD_Down
			} else {
				return elevio.MD_Up
				log.Printf("Kjøre opp")
			}
		}
		return elevio.MD_Stop
	}
}

func IsValidStop(e *elevator.Elevator) bool {
	// Check if there are any orders in the same direction
	if e.Orders[e.Floor][elevio.BT_Cab] {
		return true
	} else if e.Orders[e.Floor][elevio.BT_HallUp] && e.Dir == elevio.MD_Up {
		return true
	} else if e.Orders[e.Floor][elevio.BT_HallDown] && e.Dir == elevio.MD_Down {
		return true
	} else if e.Orders[e.Floor][elevio.BT_HallUp] && e.Dir == elevio.MD_Down && !AnyCabOrdersAhead(e) {
		return true
	} else if e.Orders[e.Floor][elevio.BT_HallDown] && e.Dir == elevio.MD_Up && !AnyCabOrdersAhead(e) {
		return true
	} else {
		return false
	}
}

// AnyCabOrdersAhead Har lyst til å lage en som ikke bruker denne
func AnyCabOrdersAhead(e *elevator.Elevator) bool {
	switch e.Dir {
	case elevio.MD_Up:
		for f := e.Floor; f < config.NumFloors; f++ {
			if e.Orders[f][elevio.BT_Cab] || e.Orders[f][elevio.BT_HallUp] {
				return true
			}
		}
		return false
	case elevio.MD_Down:
		for f := 0; f < e.Floor; f++ {
			if e.Orders[f][elevio.BT_Cab] || e.Orders[f][elevio.BT_HallDown] {
				return true
			}
		}
		return false
	default:
		return false
	}
}

// AnyOrderInDirection Burde være en finn nærmeste press knapp
func AnyOrderInDirection(e *elevator.Elevator, dir elevio.MotorDirection) bool {
	switch dir {
	case elevio.MD_Up:
		for f := e.Floor; f < config.NumFloors; f++ {
			for btn, _ := range e.Orders[f] {
				if e.Orders[f][btn] {
					return true
				}
			}
		}
		return false
	case elevio.MD_Down:
		for f := 0; f < e.Floor; f++ {
			for btn, _ := range e.Orders[f] {
				if e.Orders[f][btn] {
					return true
				}
			}
		}
		return false
	default:
		fmt.Printf("Linje 272: Her skal du ikke være \n")
		return false
	}
}

func NewStatesFromAssigner(
	newStates map[string][][3]bool,
	pid string,
	chVirtualButtons chan<- elevio.ButtonEvent,
	chRemoveOrders chan<- elevio.ButtonEvent,
	e elevator.Elevator) {
	for id, ord := range newStates {
		if id == pid {
			//fmt.Printf("New states from assigner: %+v", ord)
			for f := range ord {
				for b := range ord[f] {
					if ord[f][b] {
						//f !e.Orders[f][b] {
						chVirtualButtons <- elevio.ButtonEvent{Floor: f, Button: elevio.ButtonType(b)}
						//}
					} else {
						if e.Orders[f][b] {
							chRemoveOrders <- elevio.ButtonEvent{Floor: f, Button: elevio.ButtonType(b)}
						}
					}
				}
			}
		}
	}
}

func ClearHallOrdersAtFloor(eObj *elevator.Elevator, clearHallFsm chan<- elevio.ButtonEvent) {
	if eObj.Orders[eObj.Floor][elevio.BT_HallUp] {
		clearHallFsm <- elevio.ButtonEvent{Floor: eObj.Floor, Button: elevio.BT_HallUp}
	}
	if eObj.Orders[eObj.Floor][elevio.BT_HallDown] {
		clearHallFsm <- elevio.ButtonEvent{Floor: eObj.Floor, Button: elevio.BT_HallDown}
	}
}
