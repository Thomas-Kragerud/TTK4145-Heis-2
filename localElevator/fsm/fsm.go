package fsm

import (
	"Project/config"
	"Project/elevio"
	"Project/localElevator/elevator"
	"Project/localElevator/fsmUtils"
	"fmt"
	"log"
	"os"
	"time"
)

func Fsm(
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
	for {
		eObj.UpdateLights()
		select {
		case btnEvent := <-chAddButton:
			switch eObj.State {
			case elevator.Idle:
				if eObj.Floor == btnEvent.Floor {
					eObj.SetStateDoorOpen()
					elevio.SetDoorOpenLamp(true)
					/* doorTimer.Reset(config.DoorOpenTime) */
					chNewState <- *eObj

				} else {
					eObj.AddOrder(btnEvent) // Add order to orders
					eObj.UpdateLights()
					eObj.Dir = fsmUtils.GetNextDirection(eObj) // Find direction
					elevio.SetMotorDirection(eObj.Dir)         // Set direction
					eObj.SetStateMoving()
					chNewState <- *eObj

				}
				break

			case elevator.Moving:
				// Add order to queue
				eObj.AddOrder(btnEvent)
				
				break

			case elevator.DoorOpen:
				// Add order to queue if not on the correct floor
				if eObj.Floor == btnEvent.Floor {
					doorTimer.Reset(config.DoorOpenTime)
					//eObj.UpdateLights()

				} else {
					eObj.AddOrder(btnEvent)
				}
				//chNewState <- *eObj
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
				eObj.Dir = fsmUtils.GetNextDirection(eObj)
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

			case elevator.Moving:
				if fsmUtils.IsValidStop(eObj) {
					elevio.SetMotorDirection(elevio.MD_Stop) // Stop the elevator
					eObj.ClearOrderAtFloor(eObj.Floor)       // Clear all orders at current floor
					elevio.SetDoorOpenLamp(true)
					doorTimer.Reset(config.DoorOpenTime) // Reset the door timer
					eObj.SetStateDoorOpen()              // Set state to DoorOpen
					eObj.UpdateLights()                  // Update alle elevator lights
					chNewState <- *eObj                  // Broadcast states
				} else if (floor == 0 && eObj.Dir == elevio.MD_Down) || (floor == config.NumFloors-1 && eObj.Dir == elevio.MD_Up) || (eObj.ReAssignStop) {
					if eObj.ReAssignStop {
						eObj.ReAssignStop = false
						log.Printf("Stoppet på nærmeste nice floor\n")
					}

					eObj.Dir = elevio.MD_Stop                  // Stop elevator so it does not run out of bounds
					elevio.SetMotorDirection(eObj.Dir)         // Set direction to stop
					eObj.Dir = fsmUtils.GetNextDirection(eObj) // Find next direction
					elevio.SetMotorDirection(eObj.Dir)

					if eObj.Dir == elevio.MD_Stop {
						eObj.SetStateIdle()
						chNewState <- *eObj
					} else {
						eObj.SetStateMoving()
					}
					log.Printf("Ble redded fra å kjøre ut av bygget\n")
					log.Printf("Elevator state: %v\n", eObj.String())
				}
			default:
				log.Printf("Error: Elevator moving when it shoudnt, but received floor signal\n")
				log.Printf("Elevator state: %v\n", eObj.String())
				break
			}

		case obstruction := <-chIoObstical:
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
			//chNewState <- *eObj // Send elevator states through channel

		case stop := <-chIoStop:
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
			switch eObj.State {
			case elevator.DoorOpen:
				if eObj.Obs {
					doorTimer.Reset(config.DoorOpenTime)
					break
				}
				log.Printf("Door timer!")
				if eObj.Dir != elevio.MD_Stop {
					eObj.ClearOrderAtFloor(eObj.Floor)
					oldDir := eObj.Dir
					log.Printf("Old dir%v\n", oldDir)
					eObj.Dir = fsmUtils.GetNextDirection(eObj)
					log.Printf("Direction %v\n", eObj.Dir)
					switch eObj.Dir {
					case oldDir:
						//Close door
						elevio.SetDoorOpenLamp(false)
						eObj.SetStateMoving()
						eObj.UpdateLights()
						elevio.SetMotorDirection(eObj.Dir)
						log.Printf("Move in same direction")
						chNewState <- *eObj
						break
					
					case -oldDir:
						// Oposit direction 
						elevio.SetDoorOpenLamp(false)
						eObj.ClearOrderAtFloor(eObj.Floor) // Clear hall in oposit direction 
						eObj.SetStateMoving()
						eObj.UpdateLights()
						elevio.SetMotorDirection(eObj.Dir)
						log.Printf("Move in oposit direction")
						chNewState <- *eObj
						break
					
					case elevio.MD_Stop:
						eObj.Dir = elevio.MD_Stop //
						eObj.ClearOrderAtFloor(eObj.Floor) // Clear hall in oposit direction
						doorTimer.Reset(config.DoorOpenTime) // Reset timer 		
						break				
					}
				
				} else {
					// Motor direction is stop
					eObj.Dir = fsmUtils.GetNextDirection(eObj)
					if eObj.Dir == elevio.MD_Stop {
						eObj.SetStateIdle()
						elevio.SetDoorOpenLamp(false)
						eObj.UpdateLights()
					} else {
						eObj.ClearOrderAtFloor(eObj.Floor)
						elevio.SetDoorOpenLamp(false)
						eObj.SetStateMoving()
						elevio.SetMotorDirection(eObj.Dir)
						eObj.UpdateLights()
					}
		
				}
				
			}
		}
		log.Printf("FSM %s\n", eObj.String())
	}
}
