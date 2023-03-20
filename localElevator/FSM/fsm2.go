package FSM

import (
	"Project/config"
	"Project/elevio"
	"Project/localElevator/elevator"
	"fmt"
	"os"
	"time"
)

func FSM2(
	initElevator elevator.Elevator,
	chVirtualButtons <-chan elevio.ButtonEvent,
	chVirtualFloor <-chan int,
	chVirtualObstical <-chan bool,
	chVirtualStop <-chan bool,
	chToDist chan<- elevator.Elevator) {

	eObj := &initElevator

	// Main loop for FSM
	doorTimer := time.NewTimer(0) // Initialise timer
	for {
		eObj.UpdateLights()
		select {
		case btnEvent := <-chVirtualButtons:
			switch eObj.State {
			case elevator.Idle:
				if eObj.Floor == btnEvent.Floor {
					eObj.SetStateDoorOpen()
					elevio.SetDoorOpenLamp(true)
					doorTimer.Reset(3 * time.Second)
					chToDist <- *eObj

				} else {
					eObj.AddOrder(btnEvent)                // Add order to orders
					eObj.Dir = simple_next_direction(eObj) // Find direction
					elevio.SetMotorDirection(eObj.Dir)     // Set direction
					eObj.SetStateMoving()

					chToDist <- *eObj // Send elevator states through channel
				}
				break

			case elevator.Moving:
				// Add order to queue
				eObj.AddOrder(btnEvent)
				chToDist <- *eObj
				break

			case elevator.DoorOpen:
				// Add order to queue if not on the correct floor
				if eObj.Floor == btnEvent.Floor {
					doorTimer.Reset(3 * time.Second)
				} else {
					eObj.AddOrder(btnEvent)
					chToDist <- *eObj
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
					eObj.ClearOrderAtFloor(eObj.Floor)       // Clear all orders at current floor
					elevio.SetDoorOpenLamp(true)
					doorTimer.Reset(3 * time.Second) // Reset the door timer
					eObj.SetStateDoorOpen()          // Set state to DoorOpen
					eObj.UpdateLights()              // Update alle elevator lights
					chToDist <- *eObj                // Broadcast states
				}

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
