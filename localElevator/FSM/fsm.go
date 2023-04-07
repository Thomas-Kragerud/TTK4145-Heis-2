package FSM

import (
	"Project/config"
	"Project/elevio"
	"Project/localElevator/elevator"
	"Project/localElevator/fsm_utils"
	"fmt"
	"os"
	"time"
)

func FSM(
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
	chToDist <- *eObj
	// Main loop for FSM
	doorTimer := time.NewTimer(0) // Initialise timer
	eObj.ClearAllOrders()
	for {
		//fmt.Printf("Elevator is in state: %v\n", eObj.State)
		//fmt.Printf(" %s\n", eObj.String())
		eObj.UpdateLights()
		select {
		case btnEvent := <-chVirtualButtons:
			//fmt.Printf("**** Button event ****\n")
			//fmt.Printf(" %v\n", btnEvent)
			switch eObj.State {
			case elevator.Idle:
				if eObj.Floor == btnEvent.Floor {
					eObj.SetStateDoorOpen()
					//fmt.Printf("**** Button event ****\n")
					elevio.SetDoorOpenLamp(true)
					doorTimer.Reset(3 * time.Second)

					fsm_utils.ClearHallOrdersAtFloor(eObj, clearHallFsm)

					//chToDist <- *eObj

				} else {
					eObj.AddOrder(btnEvent)                     // Add order to orders
					eObj.Dir = fsm_utils.GetNextDirection(eObj) // Find direction
					elevio.SetMotorDirection(eObj.Dir)          // Set direction

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
					fsm_utils.ClearHallOrdersAtFloor(eObj, clearHallFsm)
					eObj.UpdateLights()

				} else {
					eObj.AddOrder(btnEvent)
					//chToDist <- *eObj
				}
				break
			}

		case msg := <-chReAssign:
			go fsm_utils.NewStatesFromAssigner(
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
					fsm_utils.ClearHallOrdersAtFloor(eObj, clearHallFsm)
					eObj.ClearOrderAtFloor(eObj.Floor) // Clear all orders at current floor
					elevio.SetDoorOpenLamp(true)

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
				fsm_utils.ClearHallOrdersAtFloor(eObj, clearHallFsm)
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

			}
			//case <-updateTimer.C:
			//	chMsgToNetwork <- *eObj
			//	updateTimer.Reset(500 * time.Millisecond)
		}
	}
}
