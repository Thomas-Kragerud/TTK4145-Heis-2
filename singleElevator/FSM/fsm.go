package FSM

import (
	"Project/config"
	"Project/singleElevator/elevator"
	"Project/singleElevator/elevio"
	"fmt"
	"os"
	"time"
)

func FSM(
	port string,
	pid string,
	chMsgToNetwork chan<- elevator.Elevator, // channel to something
	chMsgFromNetwork <-chan elevator.Elevator,
	chButtons <-chan elevio.ButtonEvent, // channel to something different
	chAtFloor <-chan int,
	chObst <-chan bool,
	chStop <-chan bool) {

	// Init elevator
	elevio.Init("localhost:"+port, 4)
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
	fmt.Printf("Nu kjør me\n")

	doorTimer := time.NewTimer(0) // Initialise timer
	//updateTimer := time.NewTimer(0)

	// Main loop for FSM
	for {
		eObj.UpdateLights()
		select {
		case btnEvent := <-chButtons:
			switch eObj.State {
			case elevator.Idle:
				if eObj.Floor == btnEvent.Floor {
					eObj.SetStateDoorOpen()
					elevio.SetDoorOpenLamp(true)
					doorTimer.Reset(3 * time.Second)
					chMsgToNetwork <- *eObj

				} else {
					eObj.AddOrder(btnEvent)                // Add order to orders
					eObj.Dir = simple_next_direction(eObj) // Find direction
					elevio.SetMotorDirection(eObj.Dir)     // Set direction
					eObj.SetStateMoving()                  // Set state moving
					// *** DISTRIBUTOR CODE ***

					chMsgToNetwork <- *eObj // Send elevator states through channel
				}
				break

			case elevator.Moving:
				// Add order to queue
				eObj.AddOrder(btnEvent)
				chMsgToNetwork <- *eObj
				break

			case elevator.DoorOpen:
				// Add order to queue if not on the correct floor
				if eObj.Floor == btnEvent.Floor {
					doorTimer.Reset(3 * time.Second)
				} else {
					eObj.AddOrder(btnEvent)
					chMsgToNetwork <- *eObj
				}
				break
			}

		case floor := <-chAtFloor:
			eObj.SetFloor(floor)
			eObj.UpdateLights()
			switch eObj.State {
			// Case Idle and Door open can not happen
			case elevator.Moving:
				//** If request say we should stop at this floor
				if valid_stop(eObj) {
					elevio.SetMotorDirection(elevio.MD_Stop) // Stop the elevator
					eObj.ClearOrderAtFloor(eObj.Floor)       // Clear all orders at current floor
					elevio.SetDoorOpenLamp(true)
					doorTimer.Reset(3 * time.Second) // Reset the door timer
					eObj.SetStateDoorOpen()          // Set state to DoorOpen
					eObj.UpdateLights()              // Update alle elevator lights
					chMsgToNetwork <- *eObj          // Broadcast states
				}

			default:
				break
			}

		case obstruction := <-chObst:
			switch eObj.State {
			case elevator.Idle:
				// Should the door not open and elevator not move?

			case elevator.Moving:
				// Should the elevator stop between floors?

			case elevator.DoorOpen:
				if obstruction {
					eObj.Obs = true
				} else {
					eObj.Obs = false
				}
			}
			chMsgToNetwork <- *eObj // Send elevator states through channel

		case stop := <-chStop:
			fmt.Printf("%+v\n", stop)
			// Clear all og exit
			for floor := 0; floor < config.NumFloors; floor++ {
				//clearOrdersAtFloor(eObj, floor)
				eObj.ClearOrderAtFloor(floor)
			}
			eObj.SetDirectionStop()
			elevio.SetMotorDirection(eObj.Dir)
			fmt.Printf(eObj.String())
			chMsgToNetwork <- *eObj // Send elevator states through channel
			os.Exit(1)

		case <-doorTimer.C:
			switch eObj.State {
			case elevator.DoorOpen:
				if eObj.Obs {
					doorTimer.Reset(3 * time.Second)
					break
				}
				eObj.Dir = simple_next_direction(eObj)
				elevio.SetMotorDirection(eObj.Dir)
				elevio.SetDoorOpenLamp(false)

				if eObj.Dir == elevio.MD_Stop {
					eObj.SetStateIdle()
					chMsgToNetwork <- *eObj
				} else {
					eObj.SetStateMoving()
				}

			}
			//case <-updateTimer.C:
			//	chMsgToNetwork <- *eObj
			//	updateTimer.Reset(500 * time.Millisecond)
		}
	}
}

func simple_next_direction(e *elevator.Elevator) elevio.MotorDirection {
	// Samme retning: Hvis orde forbi posisjon med motsatt retning
	// Samme retning; Hvis cab ordre forbi i samme rettning
	// Idle: Hvis ingen ordre
	// Motsatt retning: Hvis ordre (else)
	if e.OrderIsEmpty() {
		return elevio.MD_Stop
	} else {
		switch e.Dir {
		case elevio.MD_Up:
			for f := e.Floor; f < config.NumFloors; f++ {
				if e.Orders[f][elevio.BT_HallDown] || e.Orders[f][elevio.BT_Cab] {
					return elevio.MD_Up
				}
			}

		case elevio.MD_Down:
			for f := 0; f < e.Floor; f++ {
				if e.Orders[f][elevio.MD_Up] || e.Orders[f][elevio.BT_Cab] {
					return elevio.MD_Down
				}
			}
		case elevio.MD_Stop:
			if any_order_in_direction(e, elevio.MD_Down) {
				return elevio.MD_Down
			} else {
				return elevio.MD_Up
			}
		}
		fmt.Printf("Linje 188: Bytta rettning \n")
		return -e.Dir
	}
}

func valid_stop(e *elevator.Elevator) bool {
	// Check if there are any orders in the same direction
	if e.Orders[e.Floor][elevio.BT_Cab] {
		return true
	} else if e.Orders[e.Floor][elevio.BT_HallUp] && e.Dir == elevio.MD_Up {
		return true
	} else if e.Orders[e.Floor][elevio.BT_HallDown] && e.Dir == elevio.MD_Down {
		return true
	} else if e.Orders[e.Floor][elevio.BT_HallUp] && e.Dir == elevio.MD_Down && !cab_order_beyond(e) {
		return true
	} else if e.Orders[e.Floor][elevio.BT_HallDown] && e.Dir == elevio.MD_Up && !cab_order_beyond(e) {
		return true
	} else {
		return false
	}
}

// cab_order_beyond Har lyst til å lage en som ikke bruker denne
func cab_order_beyond(e *elevator.Elevator) bool {
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

// any_order_in_direction Burde være en finn nærmeste press knapp
func any_order_in_direction(e *elevator.Elevator, dir elevio.MotorDirection) bool {
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
