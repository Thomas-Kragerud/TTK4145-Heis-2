package FSM

import (
	"Project/config"
	elevio "Project/elevio"
	"Project/localElevator/elevator"
	"fmt"
	"os"
	"time"
)

func FSM2(
	initElevator elevator.Elevator,
	chVirtualButtons chan elevio.ButtonEvent,
	chVirtualFloor <-chan int,
	chVirtualObstical <-chan bool,
	chVirtualStop <-chan bool,
	chToDist chan<- elevator.Elevator,
	chRemoveOrders chan elevio.ButtonEvent,
	chReAssign <-chan map[string][][3]bool,
	clearHallFsm chan<- elevio.ButtonEvent) {

	c := initElevator
	eObj := &c
	// Main loop for FSM
	doorTimer := time.NewTimer(0) // Initialise timer
	eObj.ClearAllOrders()
	for {
		fmt.Printf("Elevator is in state: %v\n", eObj.State)
		fmt.Printf(" %s\n", eObj.String())
		eObj.UpdateLights()
		select {
		case btnEvent := <-chVirtualButtons:
			fmt.Printf("**** Button event ****\n")
			fmt.Printf(" %v\n", btnEvent)
			switch eObj.State {
			case elevator.Idle:
				if eObj.Floor == btnEvent.Floor {
					eObj.SetStateDoorOpen()
					fmt.Printf("**** Button event ****\n")
					elevio.SetDoorOpenLamp(true)
					doorTimer.Reset(3 * time.Second)
					if eObj.Orders[eObj.Floor][elevio.BT_HallUp] {
						clearHallFsm <- elevio.ButtonEvent{Floor: eObj.Floor, Button: elevio.BT_HallUp}
					} else if eObj.Orders[eObj.Floor][elevio.BT_HallDown] {
						clearHallFsm <- elevio.ButtonEvent{Floor: eObj.Floor, Button: elevio.BT_HallDown}
					}

					//chToDist <- *eObj

				} else {
					eObj.AddOrder(btnEvent)                // Add order to orders
					eObj.Dir = simple_next_direction(eObj) // Find direction
					elevio.SetMotorDirection(eObj.Dir)     // Set direction

					eObj.SetStateMoving()

					//chToDist <- *eObj // Send elevator states through channel
				}
				break

			case elevator.Moving:
				// Add order to queue
				eObj.AddOrder(btnEvent)
				//chToDist <- *eObj
				break

			case elevator.DoorOpen:
				// Add order to queue if not on the correct floor
				if eObj.Floor == btnEvent.Floor {
					doorTimer.Reset(3 * time.Second)
					if eObj.Orders[eObj.Floor][elevio.BT_HallUp] {
						clearHallFsm <- elevio.ButtonEvent{Floor: eObj.Floor, Button: elevio.BT_HallUp}
					} else if eObj.Orders[eObj.Floor][elevio.BT_HallDown] {
						clearHallFsm <- elevio.ButtonEvent{Floor: eObj.Floor, Button: elevio.BT_HallDown}
					}
					eObj.UpdateLights()

				} else {
					eObj.AddOrder(btnEvent)
					//chToDist <- *eObj
				}
				break
			}

		case msg := <-chReAssign:
			go newStatesFromAssigner(
				msg,
				initElevator.Id,
				chVirtualButtons,
				chRemoveOrders,
				*eObj)

		case remove := <-chRemoveOrders:
			switch eObj.State {
			case elevator.Idle:
				eObj.Orders[remove.Floor][remove.Button] = false
				eObj.UpdateLights()
				break

			case elevator.Moving:
				eObj.Orders[remove.Floor][remove.Button] = false
				eObj.UpdateLights()
				//elevio.SetMotorDirection(elevio.MD_Stop)
				eObj.Dir = simple_next_direction(eObj)
				elevio.SetMotorDirection(eObj.Dir)
				eObj.SetStateMoving()
				if eObj.Dir == elevio.MD_Stop {
					eObj.SetStateIdle()
				}
				break

			case elevator.DoorOpen:
				if eObj.Floor == remove.Floor {
					continue
				} else {
					eObj.Orders[remove.Floor][remove.Button] = false
					eObj.UpdateLights()
				}
				break

			}

		case floor := <-chVirtualFloor:
			eObj.SetFloor(floor)
			eObj.UpdateLights()
			switch eObj.State {
			// Case Idle and Door open can not happen
			case elevator.Moving:
				//** If request say we should stop at this floor
				if valid_stop(eObj) {
					elevio.SetMotorDirection(elevio.MD_Stop) // Stop the elevator
					if eObj.Orders[eObj.Floor][elevio.BT_HallUp] {
						clearHallFsm <- elevio.ButtonEvent{Floor: eObj.Floor, Button: elevio.BT_HallUp}
					} else if eObj.Orders[eObj.Floor][elevio.BT_HallDown] {
						clearHallFsm <- elevio.ButtonEvent{Floor: eObj.Floor, Button: elevio.BT_HallDown}
					}
					eObj.ClearOrderAtFloor(eObj.Floor) // Clear all orders at current floor
					elevio.SetDoorOpenLamp(true)

					doorTimer.Reset(3 * time.Second) // Reset the door timer
					eObj.SetStateDoorOpen()          // Set state to DoorOpen
					eObj.UpdateLights()              // Update alle elevator lights
					chToDist <- *eObj                // Broadcast states
				} else if (floor == 0 && eObj.Dir == elevio.MD_Down) || (floor == config.NumFloors-1 && eObj.Dir == elevio.MD_Up) {
					elevio.SetMotorDirection(elevio.MD_Stop) // Stop the elevator
					eObj.SetDirectionStop()                  // Set direction to stop
					eObj.Dir = simple_next_direction(eObj)
					elevio.SetMotorDirection(eObj.Dir)

					if eObj.Dir == elevio.MD_Stop {
						eObj.SetStateIdle()
						chToDist <- *eObj
					} else {
						eObj.SetStateMoving()
					}
				}
				break

			default:
				break
			}

		case obstruction := <-chVirtualObstical:
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
			chToDist <- *eObj // Send elevator states through channel

		case stop := <-chVirtualStop:
			fmt.Printf("%+v\n", stop)
			// Clear all og exit
			for floor := 0; floor < config.NumFloors; floor++ {
				//clearOrdersAtFloor(eObj, floor)
				eObj.ClearOrderAtFloor(floor)
			}
			eObj.SetDirectionStop()
			elevio.SetMotorDirection(eObj.Dir)
			fmt.Printf(eObj.String())
			chToDist <- *eObj // Send elevator states through channel
			os.Exit(1)

		case <-doorTimer.C:
			switch eObj.State {
			case elevator.DoorOpen:
				if eObj.Obs {
					doorTimer.Reset(3 * time.Second)
					break
				}
				if eObj.Orders[eObj.Floor][elevio.BT_HallUp] {
					clearHallFsm <- elevio.ButtonEvent{Floor: eObj.Floor, Button: elevio.BT_HallUp}
				} else if eObj.Orders[eObj.Floor][elevio.BT_HallDown] {
					clearHallFsm <- elevio.ButtonEvent{Floor: eObj.Floor, Button: elevio.BT_HallDown}
				}
				eObj.ClearOrderAtFloor(eObj.Floor) // Clear all orders at current floor
				eObj.Dir = simple_next_direction(eObj)
				elevio.SetMotorDirection(eObj.Dir)
				elevio.SetDoorOpenLamp(false)

				if eObj.Dir == elevio.MD_Stop {
					eObj.SetStateIdle()
					chToDist <- *eObj
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

func newStatesFromAssigner(
	newStates map[string][][3]bool,
	pid string,
	chVirtualButtons chan<- elevio.ButtonEvent,
	chRemoveOrders chan<- elevio.ButtonEvent,
	e elevator.Elevator) {
	for id, ord := range newStates {
		if id == pid {
			fmt.Printf("New states from assigner: %+v", ord)
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
