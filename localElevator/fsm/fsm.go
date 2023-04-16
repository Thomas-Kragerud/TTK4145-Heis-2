package fsm

import (
	"Project/config"
	"Project/elevio"
	"Project/localElevator/elevator"
	"Project/localElevator/fsm_utils"
	"fmt"
	//"log"
	"os"
	"time"
)

func FsmTest(
	eObj *elevator.Elevator,
	chIoFloor <-chan int,
	chIoObstical <-chan bool,
	chIoStop <-chan bool,
	chNewState chan<- elevator.Elevator,
	chRmButton <-chan elevio.ButtonEvent,
	chAddButton <-chan elevio.ButtonEvent,
) {
	doorTimer := time.NewTimer(0) // Initialise timer
	eObj.ClearAllOrders()
	printFSMStates := true
	for {
		eObj.UpdateLights()
		//select {
		//
		//}
		select {

		case remove := <-chRmButton:
			//log.Printf("Reasigne!!")
			if printFSMStates {fmt.Print(("FSM remove \n"))}
			switch eObj.State {
			case elevator.Idle:
				eObj.UpdateLights()
				eObj.ClearOrderFromBtn(remove)
				chNewState <- *eObj
				break

			case elevator.Moving:
				eObj.ClearOrderFromBtn(remove)
				eObj.UpdateLights()
				fmt.Print(elevio.ToStringMotorDirection( eObj.Dir))
				eObj.Dir = fsm_utils.GetNextDirection(eObj)
				elevio.SetMotorDirection(eObj.Dir)

				if eObj.Dir == elevio.MD_Stop {
					// Elevator has no orders, and is moved to the closest floor
					// Prevents elevator from stopping in between floors
					if eObj.Floor >= 0 && eObj.Floor < config.NumFloors-1 {
						elevio.SetMotorDirection(elevio.MD_Up)
					} else {
						elevio.SetMotorDirection(elevio.MD_Down)
					}
					eObj.ReAssignStop = true // Set flag to reassign stop
				}
				chNewState <- *eObj
				break

			case elevator.DoorOpen:
				if eObj.Floor == remove.Floor {
					continue
				} else {
					eObj.ClearOrderFromBtn(remove)
					eObj.UpdateLights()
				}
				//chNewState <- *eObj
				break
			}

		case btnEvent := <-chAddButton:
			if printFSMStates  {fmt.Print("FSM btnEvent \n")}
			switch eObj.State {
			case elevator.Idle:
				if eObj.Floor == btnEvent.Floor {
					eObj.SetStateDoorOpen()
					elevio.SetDoorOpenLamp(true)
					doorTimer.Reset(config.DoorOpenTime)
					eObj.AddOrder(btnEvent)

					// Save direction of hall btn
					if btnEvent.Button == elevio.BT_HallUp {
						eObj.Dir = elevio.MD_Up
					} else if btnEvent.Button == elevio.BT_HallDown {
						eObj.Dir = elevio.MD_Down
					}
					chNewState <- *eObj

				} else {
					eObj.AddOrder(btnEvent) // Add order to orders
					eObj.UpdateLights()
					eObj.Dir = fsm_utils.GetNextDirection(eObj) // Find direction
					elevio.SetMotorDirection(eObj.Dir)          // Set direction
					eObj.SetStateMoving()
					chNewState <- *eObj

				}
				break

			case elevator.Moving:
				// Add order to queue
				eObj.AddOrder(btnEvent)
				chNewState <- *eObj
				break

			case elevator.DoorOpen:
				// Add order to queue if not on the correct floor
				//log.Print("New button when door open")
				if eObj.Floor == btnEvent.Floor {
					doorTimer.Reset(config.DoorOpenTime)
					eObj.UpdateLights()
					eObj.AddOrder(btnEvent)
					chNewState <- *eObj
					//eObj.ClearOrderFromBtn(btnEvent)

				} else {
					eObj.AddOrder(btnEvent)
					chNewState <- *eObj
				}
				// Denne må clere i message handler
				
				break
			}

		case floor := <-chIoFloor:
			if printFSMStates {fmt.Print("FSM at FLoor \n")}
			eObj.SetFloor(floor)
			eObj.UpdateLights()

			fmt.Print(eObj.Orders[floor][elevio.BT_HallUp],'\n')
			if eObj.Dir == elevio.MD_Up{fmt.Print(" DIR = UP \n")}
			if eObj.Dir == elevio.MD_Stop{fmt.Print(" DIR == STOP \n")}
			switch eObj.State {	
			case elevator.Moving:
				fmt.Print("MOVING \n")
				if fsm_utils.IsValidStop(eObj) {
					fmt.Print("VALID STOP \n")
					elevio.SetMotorDirection(elevio.MD_Stop) // Stop the elevator
					eObj.ClearOrderAtFloorInDirection(eObj.Floor)       // Clear all orders at current floor
					elevio.SetDoorOpenLamp(true)
					doorTimer.Reset(config.DoorOpenTime) // Reset the door timer
					eObj.SetStateDoorOpen()              // Set state to DoorOpen
					eObj.UpdateLights()                  // Update alle elevator lights
					chNewState <- *eObj               // Broadcast states
				} else if (floor == 0 && eObj.Dir == elevio.MD_Down) || (floor == config.NumFloors-1 && eObj.Dir == elevio.MD_Up) || (eObj.ReAssignStop) {
					if eObj.ReAssignStop {
						eObj.ReAssignStop = false
						//log.Printf("Stoppet på nærmeste nice floor\n")
					}

					eObj.Dir = elevio.MD_Stop                   // Stop elevator so it does not run out of bounds
					elevio.SetMotorDirection(eObj.Dir)          // Set direction to stop
					eObj.Dir = fsm_utils.GetNextDirection(eObj) // Find next direction
					elevio.SetMotorDirection(eObj.Dir)

					if eObj.Dir == elevio.MD_Stop {
						eObj.SetStateIdle()
						chNewState <- *eObj
					} else {
						eObj.SetStateMoving()
					}
					//log.Printf("Ble redded fra å kjøre ut av bygget\n")
					//log.Printf("Elevator state: %v\n", eObj.String())
				}
			default:
				//log.Printf("Error: Elevator moving when it shoudnt, but received floor signal\n")
				//log.Printf("Elevator state: %v\n", eObj.String())
				break
			}

		case obstruction := <-chIoObstical:
			if printFSMStates {fmt.Print("FSM obs \n")}
			switch eObj.State {
			case elevator.Idle:
				// Should the door not open and elevator not move?
				if obstruction {
					eObj.Obs = true
				} else {
					eObj.Obs = false
				}

			case elevator.Moving:
				// Should the elevator stop between floors?

			case elevator.DoorOpen:
				if obstruction {
					eObj.Obs = true
				} else {
					eObj.Obs = false
				}
			}
			chNewState <- *eObj // Send elevator states through channel

		case stop := <-chIoStop:
			if printFSMStates {fmt.Print("FSM Stop \n")}
			fmt.Printf("%+v\n", stop)
			for floor := 0; floor < config.NumFloors; floor++ {
				eObj.ClearOrderAtFloor(floor)
			}
			eObj.Dir = elevio.MD_Stop
			elevio.SetMotorDirection(eObj.Dir)
			eObj.ClearAllOrders()
			chNewState <- *eObj // Send elevator states through channel
			os.Exit(1)

		case <-doorTimer.C:
			if printFSMStates {fmt.Print(" FSM doortimer \n")}
			switch eObj.State {
			case elevator.DoorOpen:
				if eObj.Obs {
					doorTimer.Reset(config.DoorOpenTime)
					break
				}
				//log.Printf("Door timer!")
				if eObj.Dir != elevio.MD_Stop {
					fmt.Print(" YEET YEET \n")

					eObj.ClearOrderAtFloorInDirection(eObj.Floor)
					oldDir := eObj.Dir
					//log.Printf("Old dir%v\n", oldDir)
					eObj.Dir = fsm_utils.GetNextDirection(eObj)
					//log.Printf("Direction %v\n", eObj.Dir)
					switch eObj.Dir {
					case oldDir:
						//Close door
						elevio.SetDoorOpenLamp(false)
						eObj.SetStateMoving()
						eObj.UpdateLights()
						elevio.SetMotorDirection(eObj.Dir)
						//log.Printf("Move in same direction")
						chNewState <- *eObj
						break
					
					case -oldDir:
						// Oposit direction 
						elevio.SetDoorOpenLamp(false)
						eObj.ClearOrderAtFloor(eObj.Floor) // Clear hall in oposit direction 
						eObj.SetStateMoving()
						eObj.UpdateLights()
						elevio.SetMotorDirection(eObj.Dir)
						//log.Printf("Move in oposit direction")
						chNewState <- *eObj
						break
					
					case elevio.MD_Stop:
						eObj.Dir = elevio.MD_Stop //
						eObj.ClearOrderAtFloor(eObj.Floor) // Clear hall in oposit direction
						doorTimer.Reset(config.DoorOpenTime) // Reset timer 		
						break				
					}
				
				} else {
					fmt.Print(" PENIS PENIS \n")
					// Motor direction is stop
					eObj.Dir = fsm_utils.GetNextDirection(eObj)
					if eObj.Dir == elevio.MD_Stop {
						eObj.SetStateIdle()
						elevio.SetDoorOpenLamp(false)
						eObj.UpdateLights()
						chNewState <- *eObj
					} else {
						eObj.ClearOrderAtFloorInDirection(eObj.Floor)
						elevio.SetDoorOpenLamp(false)
						eObj.SetStateMoving()
						elevio.SetMotorDirection(eObj.Dir)
						eObj.UpdateLights()
						chNewState <- *eObj
					}
		
				}
				chNewState <- *eObj
			}
		}
	}
}
