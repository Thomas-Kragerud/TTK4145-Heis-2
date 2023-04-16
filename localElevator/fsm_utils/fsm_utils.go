package fsm_utils

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
			// Try floor -1
			for f := e.Floor + 1; f < config.NumFloors; f++ {
				if e.Orders[f][elevio.BT_HallDown] || e.Orders[f][elevio.BT_Cab] && (e.Floor != config.NumFloors-1) {
					return elevio.MD_Up
				}
			}

		case elevio.MD_Down:
			for f := 0; f < e.Floor; f++ {
				if e.Orders[f][elevio.BT_HallUp] || e.Orders[f][elevio.BT_Cab] && (e.Floor != 0) {
					return elevio.MD_Down
				}
			}
		case elevio.MD_Stop:
			if AnyOrderInDirection(e, elevio.MD_Down) {
				// Any order below?
				return elevio.MD_Down
			} else if AnyOrderInDirection(e, elevio.MD_Up) {
				// Any order above?
				return elevio.MD_Up
			} else {
				// No orders except own floor
				return elevio.MD_Stop
			}
		}
		fmt.Printf("Linje 188: Bytta rettning \n")
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
		for f := e.Floor + 1; f < config.NumFloors; f++ {
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
		log.Fatalf("Var i default i AnyCabOrdersAhead %s", e.String())
		return false
	}
}

// AnyOrderInDirection Burde være en finn nærmeste press knapp
func AnyOrderInDirection(e *elevator.Elevator, dir elevio.MotorDirection) bool {
	switch dir {
	case elevio.MD_Up:
		for f := e.Floor + 1; f < config.NumFloors; f++ {
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

func ClearOrderInDirection(Orders [][]bool, floor int, Dir elevio.MotorDirection) ([][]bool, *elevio.ButtonEvent, *elevio.ButtonEvent) {
	var hallbtn *elevio.ButtonEvent
	var cabbtn *elevio.ButtonEvent
	ord := make([][]bool, len(Orders))
	copy(ord, Orders)

	switch Dir {
	case elevio.MD_Up:
		if ord[floor][elevio.BT_HallUp] {
			Orders[floor][elevio.BT_HallUp] = false
			hallbtn = &elevio.ButtonEvent{floor, elevio.BT_HallUp}
		} else if ord[floor][elevio.BT_HallDown] && floor == config.NumFloors-1 {
			Orders[floor][elevio.BT_HallDown] = false
			hallbtn = &elevio.ButtonEvent{floor, elevio.BT_HallDown}
			log.Printf("Skal nu i clear ord")
		}
	case elevio.MD_Down:
		if ord[floor][elevio.BT_HallDown] {
			Orders[floor][elevio.BT_HallDown] = false
			hallbtn = &elevio.ButtonEvent{floor, elevio.BT_HallDown}
		} else if ord[floor][elevio.BT_HallUp] && floor == 0 {
			Orders[floor][elevio.BT_HallUp] = false
			hallbtn = &elevio.ButtonEvent{floor, elevio.BT_HallUp}
			log.Printf("Skal nu i clear ord nede")
		}
	}

	if ord[floor][elevio.BT_Cab] {
		ord[floor][elevio.BT_Cab] = false
		cabbtn = &elevio.ButtonEvent{floor, elevio.BT_Cab}
	}

	log.Printf("Sendte buttons")
	return ord, hallbtn, cabbtn
}

func ClearOrderWhenMDStop(Orders [][]bool, floor int) ([][]bool, *elevio.ButtonEvent, *elevio.ButtonEvent) {
	var hallbtn *elevio.ButtonEvent
	var cabbtn *elevio.ButtonEvent
	ord := make([][]bool, len(Orders))
	copy(ord, Orders)

	if ord[floor][elevio.BT_HallUp] && ord[floor][elevio.BT_HallDown] {
		log.Fatalf("Error ClearOrderWhenMDStop - Skal ikke være begge knappene trykket samtidig")
	}

	if ord[floor][elevio.BT_HallUp] {
		Orders[floor][elevio.BT_HallUp] = false
		hallbtn = &elevio.ButtonEvent{floor, elevio.BT_HallUp}

	} else if ord[floor][elevio.BT_HallDown] {
		Orders[floor][elevio.BT_HallDown] = false
		hallbtn = &elevio.ButtonEvent{floor, elevio.BT_HallDown}

	}

	if ord[floor][elevio.BT_Cab] {
		ord[floor][elevio.BT_Cab] = false
		cabbtn = &elevio.ButtonEvent{floor, elevio.BT_Cab}
	}

	return ord, hallbtn, cabbtn
}
