package FSM2

import (
	"Project/config"
	"Project/elevio"
	"Project/localElevator/elevator"
	"Project/localElevator/fsm_utils"
	"Project/sound"
	"fmt"
	"os"
	"time"
)

func FsmTest(
	initElevator elevator.Elevator,
	chVirtualFloor <-chan int,
	chVirtualObstical <-chan bool,
	chVirtualStop <-chan bool,
	chToDist chan<- elevator.Elevator,
	chRmButton <-chan elevio.ButtonEvent,
	chAddButton <-chan elevio.ButtonEvent,

) {

	c := initElevator
	eObj := &c
	chToDist <- *eObj

	doorTimer := time.NewTimer(0) // Initialise timer
	eObj.ClearAllOrders()
	for {
		eObj.UpdateLights()
		select {
		case btnEvent := <-chAddButton:

			switch eObj.State {
			case elevator.Idle:
				if eObj.Floor == btnEvent.Floor {
					eObj.SetStateDoorOpen()
					elevio.SetDoorOpenLamp(true)
					doorTimer.Reset(3 * time.Second)
					chToDist <- *eObj

				} else {
					eObj.AddOrder(btnEvent)                     // Add order to orders
					eObj.Dir = fsm_utils.GetNextDirection(eObj) // Find direction
					elevio.SetMotorDirection(eObj.Dir)          // Set direction
					eObj.SetStateMoving()
					chToDist <- *eObj
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
					eObj.UpdateLights()

				} else {
					eObj.AddOrder(btnEvent)

				}
				chToDist <- *eObj
				break
			}

		case remove := <-chRmButton:
			switch eObj.State {
			case elevator.Idle:
				eObj.Orders[remove.Floor][remove.Button] = false
				eObj.UpdateLights()
				break

			case elevator.Moving:
				eObj.Orders[remove.Floor][remove.Button] = false
				eObj.UpdateLights()
				//elevio.SetMotorDirection(elevio.MD_Stop)
				eObj.Dir = fsm_utils.GetNextDirection(eObj)
				elevio.SetMotorDirection(eObj.Dir)

				if eObj.Dir == elevio.MD_Stop {
					eObj.SetStateIdle()
				} else {
					eObj.SetStateMoving()
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
				if fsm_utils.IsValidStop(eObj) {
					elevio.SetMotorDirection(elevio.MD_Stop) // Stop the elevator
					eObj.ClearOrderAtFloor(eObj.Floor)       // Clear all orders at current floor
					elevio.SetDoorOpenLamp(true)
					go sound.AtFloor(floor)
					doorTimer.Reset(3 * time.Second) // Reset the door timer
					eObj.SetStateDoorOpen()          // Set state to DoorOpen
					eObj.UpdateLights()              // Update alle elevator lights
					chToDist <- *eObj                // Broadcast states
				} else if (floor == 0 && eObj.Dir == elevio.MD_Down) || (floor == config.NumFloors-1 && eObj.Dir == elevio.MD_Up) {
					elevio.SetMotorDirection(elevio.MD_Stop) // Stop the elevator
					eObj.SetDirectionStop()                  // Set direction to stop
					eObj.Dir = fsm_utils.GetNextDirection(eObj)
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
			//fmt.Printf(eObj.String())
			chToDist <- *eObj // Send elevator states through channel
			os.Exit(1)

		case <-doorTimer.C:
			switch eObj.State {
			case elevator.DoorOpen:
				if eObj.Obs {
					doorTimer.Reset(3 * time.Second)
					break
				}
				eObj.ClearOrderAtFloor(eObj.Floor) // Clear all orders at current floor
				eObj.Dir = fsm_utils.GetNextDirection(eObj)
				elevio.SetMotorDirection(eObj.Dir)
				elevio.SetDoorOpenLamp(false)

				if eObj.Dir == elevio.MD_Stop {
					eObj.SetStateIdle()
					chToDist <- *eObj
				} else {
					eObj.SetStateMoving()
				}
				chToDist <- *eObj
			}
			//case <-updateTimer.C:
			//	chMsgToNetwork <- *eObj
			//	updateTimer.Reset(500 * time.Millisecond)
		}
	}
}
