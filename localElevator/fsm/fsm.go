package fsm

import (
	"Project/config"
	"Project/elevio"
	"Project/localElevator/elevator"
	"Project/localElevator/fsm_utils"
	"fmt"
	"log"
	"os"
	"time"
)

func FsmTest(
	eObj *elevator.Elevator,
	chIoFloor <-chan int,
	chIoObstical <-chan bool,
	chIoStop <-chan bool,
	chStateUpdate chan<- FsmOutput,
	chRmButton <-chan elevio.ButtonEvent,
	chAddButton <-chan elevio.ButtonEvent,
) {
	doorTimer := time.NewTimer(0) // Initialise timer
	stuckTimer := time.NewTimer(0)
	stuck := false
	eObj.ClearAllOrders()
	iter := 0
	for {
		eObj.UpdateLights()
		log.Printf("Start FSM iteration %d\n", iter)
		select {
		case btnEvent := <-chAddButton:
			switch eObj.State {
			case elevator.Idle:
				if eObj.Floor == btnEvent.Floor {
					eObj.SetStateDoorOpen()
					elevio.SetDoorOpenLamp(true)
					doorTimer.Reset(config.DoorOpenTime)
					//eObj.AddOrder(btnEvent) // Add order to orders - Do not need it really, but for consistency

					if btnEvent.Button != elevio.BT_Cab {
						chStateUpdate <- FsmOutput{
							Elevator: *eObj,
							Event:    ClearHall,
							BtnEvent: btnEvent,
						}
						if btnEvent.Button == elevio.BT_HallUp {
							eObj.Dir = elevio.MD_Up
						} else {
							eObj.Dir = elevio.MD_Down
						}
					} else {
						chStateUpdate <- FsmOutput{
							Elevator: *eObj,
							Event:    ClearCab,
							BtnEvent: btnEvent,
						}
					}

				} else {
					eObj.AddOrder(btnEvent) // Add order to orders
					eObj.UpdateLights()
					//************
					eObj.Dir = fsm_utils.GetNextDirection(eObj) // Find direction, skal den være der 'her'?
					elevio.SetMotorDirection(eObj.Dir)          // Set direction
					eObj.SetStateMoving()
					chStateUpdate <- FsmOutput{
						Elevator: *eObj,
						Event:    Update,
						BtnEvent: btnEvent,
					}
				}
				break

			case elevator.Moving:
				// Add order to queue
				eObj.AddOrder(btnEvent)
				chStateUpdate <- FsmOutput{
					Elevator: *eObj,
					Event:    Update,
					BtnEvent: btnEvent,
				}
				break

			case elevator.DoorOpen:
				// Add order to queue if not on the correct floor
				if eObj.Floor == btnEvent.Floor {
					doorTimer.Reset(config.DoorOpenTime)
					eObj.UpdateLights()
					if btnEvent.Button == elevio.BT_Cab {
						chStateUpdate <- FsmOutput{
							Elevator: *eObj,
							Event:    ClearCab,
							BtnEvent: btnEvent,
						}
					}
					if btnEvent.Button == elevio.BT_HallUp && eObj.Dir == elevio.MD_Up {
						chStateUpdate <- FsmOutput{
							Elevator: *eObj,
							Event:    ClearHall,
							BtnEvent: btnEvent,
						}
					} else if btnEvent.Button == elevio.BT_HallDown && eObj.Dir == elevio.MD_Down {
						chStateUpdate <- FsmOutput{
							Elevator: *eObj,
							Event:    ClearHall,
							BtnEvent: btnEvent,
						}
					} else if btnEvent.Button != elevio.BT_Cab && eObj.Dir == elevio.MD_Stop {
						chStateUpdate <- FsmOutput{
							Elevator: *eObj,
							Event:    ClearHall,
							BtnEvent: btnEvent,
						}
						if btnEvent.Button == elevio.BT_HallUp {
							eObj.Dir = elevio.MD_Up

						} else {
							eObj.Dir = elevio.MD_Down

						}
					} else {
						eObj.AddOrder(btnEvent)
						chStateUpdate <- FsmOutput{
							Elevator: *eObj,
							Event:    Update,
							BtnEvent: btnEvent,
						}

					}
				} else {
					eObj.AddOrder(btnEvent)
					chStateUpdate <- FsmOutput{
						Elevator: *eObj,
						Event:    Update,
						BtnEvent: btnEvent,
					}
				}
				break
			}

		case remove := <-chRmButton:
			switch eObj.State {
			case elevator.Idle:
				eObj.UpdateLights()
				chStateUpdate <- FsmOutput{
					Elevator: *eObj,
					Event:    Update,
				}
				break

			case elevator.Moving:
				eObj.ClearOrderFromBtn(remove)
				eObj.UpdateLights()
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
				chStateUpdate <- FsmOutput{
					Elevator: *eObj,
					Event:    Update,
				}
				break

			case elevator.DoorOpen:
				if eObj.Floor == remove.Floor {
					eObj.ClearOrderFromBtn(remove)
					continue
				} else {
					eObj.ClearOrderFromBtn(remove)
					eObj.UpdateLights()
				}
				chStateUpdate <- FsmOutput{
					Elevator: *eObj,
					Event:    Update,
				}
				break
			}

		case floor := <-chIoFloor:
			if stuck {
				stuck = false
				eObj.Obs = false
				chStateUpdate <- FsmOutput{
					Event:    ClearedObstruction,
					Elevator: *eObj,
				}
			}
			eObj.SetFloor(floor)
			eObj.UpdateLights()
			stuckTimer.Reset(4 * time.Second)

			switch eObj.State {

			case elevator.Moving:
				if fsm_utils.IsValidStop(eObj) {
					elevio.SetMotorDirection(elevio.MD_Stop) // Stop the elevator
					//eObj.ClearOrderAtFloorInDirection(eObj.Floor) // Clear all orders at current floor
					var hallBtn, cabBtn *elevio.ButtonEvent
					eObj.Orders, hallBtn, cabBtn = fsm_utils.ClearOrderInDirection(eObj.Orders, eObj.Floor, eObj.Dir)
					log.Printf("Hallbtn: %v, Cabbtn: %v \n", hallBtn, cabBtn)
					elevio.SetDoorOpenLamp(true)
					doorTimer.Reset(config.DoorOpenTime) // Reset the door timer
					eObj.SetStateDoorOpen()              // Set state to DoorOpen
					eObj.UpdateLights()                  // Update alle elevator lights

					if hallBtn != nil {
						chStateUpdate <- FsmOutput{
							Elevator: *eObj,
							Event:    ClearHall,
							BtnEvent: *hallBtn,
						}
						log.Printf("Sendte btn event")
					}
					if cabBtn != nil {
						chStateUpdate <- FsmOutput{
							Elevator: *eObj,
							Event:    ClearCab,
							BtnEvent: *cabBtn,
						}
					}

				} else if (floor == 0 && eObj.Dir == elevio.MD_Down) || (floor == config.NumFloors-1 && eObj.Dir == elevio.MD_Up) || (eObj.ReAssignStop) {
					if eObj.ReAssignStop {
						eObj.ReAssignStop = false
						log.Printf("Stoppet på nærmeste nice floor\n")
					}

					eObj.Dir = elevio.MD_Stop                   // Stop elevator so it does not run out of bounds
					elevio.SetMotorDirection(eObj.Dir)          // Set direction to stop
					eObj.Dir = fsm_utils.GetNextDirection(eObj) // Find next direction
					elevio.SetMotorDirection(eObj.Dir)

					if eObj.Dir == elevio.MD_Stop {
						eObj.SetStateIdle()
						chStateUpdate <- FsmOutput{
							Elevator: *eObj,
							Event:    Update,
						}
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
					chStateUpdate <- FsmOutput{
						Event:    Obstruction,
						Elevator: *eObj,
					}
				} else {
					eObj.Obs = false
					doorTimer.Reset(config.DoorOpenTime)
					chStateUpdate <- FsmOutput{
						Event:    ClearedObstruction,
						Elevator: *eObj,
					}
				}
			}
		case <-chIoStop:
			os.Exit(1)

		case <-doorTimer.C:
			stuckTimer.Reset(4*time.Second)
			if stuck {
				stuck = false
				eObj.Obs = false
				chStateUpdate <- FsmOutput{
					Event:    ClearedObstruction,
					Elevator: *eObj,
				}
			}
			switch eObj.State {
			case elevator.DoorOpen:
				fmt.Printf("Door timer!\n")
				if eObj.Obs {
					doorTimer.Reset(config.DoorOpenTime)
					break
				}
				var hallBtn, cabBtn *elevio.ButtonEvent
				if eObj.Dir != elevio.MD_Stop {
					// Was idle and now door is open
					//eObj.Dir = fsm_utils.GetNextDirection(eObj) // Still do not know which btn to clear

					eObj.Orders, hallBtn, cabBtn = fsm_utils.ClearOrderInDirection(eObj.Orders, eObj.Floor, eObj.Dir)
					if hallBtn != nil {
						chStateUpdate <- FsmOutput{
							Elevator: *eObj,
							Event:    ClearHall,
							BtnEvent: *hallBtn,
						}
						log.Printf("Sendte btn event")
					}
					if cabBtn != nil {
						chStateUpdate <- FsmOutput{
							Elevator: *eObj,
							Event:    ClearCab,
							BtnEvent: *cabBtn,
						}
					}
					prevDir := eObj.Dir
					switch fsm_utils.GetNextDirection(eObj) {
					case prevDir:
						// Same direction
						log.Printf("Same direction\n")
						elevio.SetDoorOpenLamp(false)
						eObj.SetStateMoving()
						elevio.SetMotorDirection(eObj.Dir)
						eObj.UpdateLights()
						chStateUpdate <- FsmOutput{
							Elevator: *eObj,
							Event:    Update,
						}
						break

					case -prevDir:
						log.Printf("Go Oposit direction\n")
						elevio.SetDoorOpenLamp(false)
						eObj.Dir = -eObj.Dir
						eObj.Orders, hallBtn, cabBtn = fsm_utils.ClearOrderInDirection(eObj.Orders, eObj.Floor, eObj.Dir)
						eObj.SetStateMoving()
						elevio.SetMotorDirection(eObj.Dir)
						if hallBtn != nil {
							chStateUpdate <- FsmOutput{
								Elevator: *eObj,
								Event:    ClearHall,
								BtnEvent: *hallBtn,
							}
							log.Printf("Sendte btn event")
						}
						if cabBtn != nil {
							chStateUpdate <- FsmOutput{
								Elevator: *eObj,
								Event:    ClearCab,
								BtnEvent: *cabBtn,
							}
						}
						eObj.UpdateLights()
						break

					case elevio.MD_Stop:
						//eObj.Dir = elevio.MD_Stop
						if eObj.OrderIsEmpty() {
							eObj.Dir = elevio.MD_Stop
							log.Printf("No orders\n")
							elevio.SetDoorOpenLamp(false)
							eObj.SetStateIdle()
							chStateUpdate <- FsmOutput{
								Elevator: *eObj,
								Event:    Update,
							}
							break
						} else {
							eObj.Dir = elevio.MD_Stop
							log.Printf("Switiching Direction\n")
							doorTimer.Reset(config.DoorOpenTime)
							//eObj.ClearOrderAtFloor(eObj.Floor)
							eObj.Orders, hallBtn, cabBtn = fsm_utils.ClearOrderWhenMDStop(eObj.Orders, eObj.Floor)
							if hallBtn != nil {
								chStateUpdate <- FsmOutput{
									Elevator: *eObj,
									Event:    ClearHall,
									BtnEvent: *hallBtn,
								}
								log.Printf("Sendte btn event")
							}
							if cabBtn != nil {
								chStateUpdate <- FsmOutput{
									Elevator: *eObj,
									Event:    ClearCab,
									BtnEvent: *cabBtn,
								}
							}
							break
						}
					}
				} else {
					// No direction
					log.Printf("Dir stop\n")
					eObj.Dir = fsm_utils.GetNextDirection(eObj)
					if eObj.Dir == elevio.MD_Stop {
						eObj.SetStateIdle()
						elevio.SetDoorOpenLamp(false)
						eObj.UpdateLights()
						chStateUpdate <- FsmOutput{
							Elevator: *eObj,
							Event:    Update,
						}
					} else {
						//eObj.ClearOrderAtFloorInDirection(eObj.Floor)
						eObj.Orders, hallBtn, cabBtn = fsm_utils.ClearOrderInDirection(eObj.Orders, eObj.Floor, eObj.Dir)

						elevio.SetDoorOpenLamp(false)
						eObj.SetStateMoving()
						elevio.SetMotorDirection(eObj.Dir)
						eObj.UpdateLights()
						if hallBtn != nil {
							chStateUpdate <- FsmOutput{
								Elevator: *eObj,
								Event:    ClearHall,
								BtnEvent: *hallBtn,
							}
							log.Printf("Sendte btn event")
						}
						if cabBtn != nil {
							chStateUpdate <- FsmOutput{
								Elevator: *eObj,
								Event:    ClearCab,
								BtnEvent: *cabBtn,
							}
						}
					}
				}
			}
		case <-stuckTimer.C:
			if eObj.State == elevator.Idle{
				stuckTimer.Reset(4 * time.Second)
			} else {
				eObj.Obs = true
				stuck = true
					chStateUpdate <- FsmOutput{
						Event:    Obstruction,
						Elevator: *eObj,
					}
			}
			


		}
		time.Sleep(config.PollRate)
		log.Printf("End of FSM iteration %d\n", iter)
		fmt.Printf("\n%s\n", eObj.String())
		iter++
		log.Printf("Direction (post) : %v\n", eObj.Dir)
	}
}
