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
	initElevator *elevator.Elevator,
	chIoFloor <-chan int,
	chIoObstical <-chan bool,
	chIoStop <-chan bool,
	chStateUpdate chan<- elevator.Elevator,
	chRmButton <-chan elevio.ButtonEvent,
	chAddButton <-chan elevio.ButtonEvent,

) {

	//eCopy := initElevator
	//eObj := &eCopy
	//chStateUpdate <- *eObj        // Broadcast init state
	eObj := initElevator
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
					chStateUpdate <- *eObj

				} else {
					eObj.AddOrder(btnEvent) // Add order to orders
					eObj.UpdateLights()
					eObj.Dir = fsm_utils.GetNextDirection(eObj) // Find direction
					elevio.SetMotorDirection(eObj.Dir)          // Set direction
					eObj.SetStateMoving()
					chStateUpdate <- *eObj
				}
				break

			case elevator.Moving:
				// Add order to queue
				eObj.AddOrder(btnEvent)
				//chStateUpdate <- *eObj
				break

			case elevator.DoorOpen:
				// Add order to queue if not on the correct floor
				if eObj.Floor == btnEvent.Floor {
					doorTimer.Reset(3 * time.Second)
					eObj.UpdateLights()

				} else {
					eObj.AddOrder(btnEvent)

				}
				chStateUpdate <- *eObj
				break
			}

		case remove := <-chRmButton:
			switch eObj.State {
			case elevator.Idle:
				eObj.UpdateLights()
				break

			case elevator.Moving:
				eObj.ClearOrderFromBtn(remove)
				eObj.UpdateLights()
				eObj.Dir = fsm_utils.GetNextDirection(eObj)
				elevio.SetMotorDirection(eObj.Dir)

				if eObj.Dir == elevio.MD_Stop {
					// Move to closes stable floor
					if eObj.Floor >= 0 && eObj.Floor < config.NumFloors-1 {
						elevio.SetMotorDirection(elevio.MD_Up)
					} else {
						elevio.SetMotorDirection(elevio.MD_Down)
					}
					eObj.SetStateIdle()
				} else {
					eObj.SetStateMoving()
				}

				break

			case elevator.DoorOpen:
				if eObj.Floor == remove.Floor {
					continue
				} else {
					eObj.ClearOrderFromBtn(remove)
					eObj.UpdateLights()
				}
				break

			}

		case floor := <-chIoFloor:
			eObj.SetFloor(floor)
			eObj.UpdateLights()
			switch eObj.State {
			// Case Idle and Door open can not happen

			case elevator.Idle:
				// Special case where elevator has no orders and would
				// otherwise be stuck between to floors due to reAssigning
				eObj.Dir = elevio.MD_Stop
				elevio.SetMotorDirection(eObj.Dir)

			case elevator.Moving:
				if fsm_utils.IsValidStop(eObj) {
					elevio.SetMotorDirection(elevio.MD_Stop) // Stop the elevator
					eObj.ClearOrderAtFloor(eObj.Floor)       // Clear all orders at current floor
					elevio.SetDoorOpenLamp(true)
					go sound.AtFloor(floor)          // Announce the floor through the speaker
					doorTimer.Reset(3 * time.Second) // Reset the door timer
					eObj.SetStateDoorOpen()          // Set state to DoorOpen
					eObj.UpdateLights()              // Update alle elevator lights
					chStateUpdate <- *eObj           // Broadcast states
				} else if (floor == 0 && eObj.Dir == elevio.MD_Down) || (floor == config.NumFloors-1 && eObj.Dir == elevio.MD_Up) {
					// Stop elevator so it does not run out of bounds
					eObj.Dir = elevio.MD_Stop
					elevio.SetMotorDirection(eObj.Dir)          // Set direction to stop
					eObj.Dir = fsm_utils.GetNextDirection(eObj) // Find next direction
					elevio.SetMotorDirection(eObj.Dir)

					if eObj.Dir == elevio.MD_Stop {
						eObj.SetStateIdle()
						chStateUpdate <- *eObj
					} else {
						eObj.SetStateMoving()
					}
				}
				break

			default:
				break
			}

		case obstruction := <-chIoObstical:
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
			chStateUpdate <- *eObj // Send elevator states through channel

		case stop := <-chIoStop:
			fmt.Printf("%+v\n", stop)
			// Clear all og exit
			for floor := 0; floor < config.NumFloors; floor++ {
				//clearOrdersAtFloor(eObj, floor)
				eObj.ClearOrderAtFloor(floor)
			}
			//eObj.SetDirectionStop()
			eObj.Dir = elevio.MD_Stop
			elevio.SetMotorDirection(eObj.Dir)
			//fmt.Printf(eObj.String())
			chStateUpdate <- *eObj // Send elevator states through channel
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
					chStateUpdate <- *eObj
				} else {
					eObj.SetStateMoving()
				}
				chStateUpdate <- *eObj
			}
			//case <-updateTimer.C:
			//	chMsgToNetwork <- *eObj
			//	updateTimer.Reset(500 * time.Millisecond)
		}
	}
}
