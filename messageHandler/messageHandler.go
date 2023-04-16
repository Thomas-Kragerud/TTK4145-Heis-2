package messageHandler

import (
	"Project/config"
	"Project/elevio"
	"Project/localElevator/elevator"
	"Project/localElevator/fsm"
	"Project/network/peers"
	"fmt"
	"log"
	"time"
)

func Handel(
	elevator *elevator.Elevator,
	chIoButtons <-chan elevio.ButtonEvent,
	chMsgFromNetwork <-chan NetworkPackage,
	chMsgToNetwork chan<- NetworkPackage,
	chFromFsm <-chan fsm.FsmOutput,
	chAddButtonToFsm chan<- elevio.ButtonEvent,
	chRmButtonFromFsm chan<- elevio.ButtonEvent,
	chPeerUpdate <-chan peers.PeerUpdate,
) {

	thisElev := elevator
	elevatorMap := make(map[string]ElevatorUpdate)
	elevatorMap[thisElev.Id] = ElevatorUpdate{*thisElev, true}
	hall := make([][2]bool, config.NumFloors)
	//reRunRate := 2000 * time.Millisecond
	//reRunTimer := time.NewTimer(10 * time.Second)

	// Anonymous function that handles the sending to the fsm
	sendToFsm := func(fromReAssigner []assignValue) {
		for _, val := range fromReAssigner {
			time.Sleep(config.PollRate)
			if val.Type == Remove {
				chRmButtonFromFsm <- val.BtnEvent
			} else if val.Type == Add {
				chAddButtonToFsm <- val.BtnEvent
			}
		}
	}
	chMsgToNetwork <- NetworkPackage{
		Event:    UpdateElevState,
		Elevator: *thisElev,
	}

	for {
		select {
		case ioButton := <-chIoButtons:
			if ioButton.Button == elevio.BT_Cab {
				e := elevatorMap[thisElev.Id]
				e.Elevator.AddOrder(ioButton)
				elevatorMap[thisElev.Id] = e
				chAddButtonToFsm <- ioButton
				fromReAssigner, err := reAssign(thisElev.Id, elevatorMap, hall)
				if err != nil {
					log.Print("None fatal error: \n", err)
				} else {
					sendToFsm(fromReAssigner)
				}
				chMsgToNetwork <- NetworkPackage{
					NewCab,
					e.Elevator,
					ioButton,
				}
			} else {
				e := elevatorMap[thisElev.Id]
				hall = addHallBTN(hall, ioButton) // Add hall to this elevator list of hall
				updateHallLights(hall)

				msg := NetworkPackage{
					NewHall,
					e.Elevator,
					ioButton,
				}
				chMsgToNetwork <- msg

				fromReAssigner, err := reAssign(thisElev.Id, elevatorMap, hall)
				if err != nil {
					log.Print("None fatal error: \n", err)
				} else {
					sendToFsm(fromReAssigner)
				}
			}

		case updateStateFromFsm := <-chFromFsm:
			// New information about the local elevator from the fsm
			updateElevator := updateStateFromFsm.Elevator
			e := elevatorMap[thisElev.Id]
			e.Elevator = updateElevator
			elevatorMap[thisElev.Id] = e

			switch updateStateFromFsm.Event {
			case fsm.ClearHall:
				hall = clareHallBTN(hall, updateStateFromFsm.BtnEvent)
				updateHallLights(hall)
				chMsgToNetwork <- NetworkPackage{
					Event:    ClareHall,
					Elevator: updateStateFromFsm.Elevator,
					BtnEvent: updateStateFromFsm.BtnEvent,
				}
			case fsm.ClearCab:
				chMsgToNetwork <- NetworkPackage{
					Event:    UpdateElevState,
					Elevator: updateStateFromFsm.Elevator,
					BtnEvent: updateStateFromFsm.BtnEvent,
				}

			case fsm.Update:
				chMsgToNetwork <- NetworkPackage{
					Event:    UpdateElevState,
					Elevator: updateStateFromFsm.Elevator,
				}
			}

		case msgFromNet := <-chMsgFromNetwork:
			if msgFromNet.Elevator.Id == thisElev.Id {
				if msgFromNet.Event == Recover {
					// *****
					for f := 0; f < config.NumFloors; f++ {
						for b := elevio.ButtonType(0); b < 3; b++ {
							if msgFromNet.Elevator.Orders[f][b] {
								time.Sleep(config.PollRate)
								chAddButtonToFsm <- elevio.ButtonEvent{f, b}
							}
						}
					}
					chMsgToNetwork <- NetworkPackage{
						Event:    RecoveredElevator,
						Elevator: msgFromNet.Elevator,
					}
					break
				}
			} else if _, ok := elevatorMap[msgFromNet.Elevator.Id]; !ok {
				if msgFromNet.Elevator.Id != "" {
					// Have not seen this elevator before
					newElevator := ElevatorUpdate{msgFromNet.Elevator, true}
					elevatorMap[msgFromNet.Elevator.Id] = newElevator
					this := elevatorMap[thisElev.Id] // Broadcast to net so the new elevator see the other elevators on the network
					chMsgToNetwork <- NetworkPackage{
						Event:    UpdateElevState,
						Elevator: this.Elevator,
					}
				} else {
					log.Printf("Hallisinuerte en heis\n")
				}
			} else if e, ok := elevatorMap[msgFromNet.Elevator.Id]; ok && !e.Alive {
				// Recover dead elevator that is now back online
				elevatorMap[e.Elevator.Id] = e // Extract old elevator
				e.Alive = true                 // Set alive
				elevatorMap[e.Elevator.Id] = e // Update elevator map
				// Broadcast to net so elevator can retrieve its old state
				chMsgToNetwork <- NetworkPackage{
					Event:    Recover,
					Elevator: e.Elevator,
				}
				break

			} else {
				// Update elevator state
				e := elevatorMap[msgFromNet.Elevator.Id]
				e.Elevator = msgFromNet.Elevator
				elevatorMap[msgFromNet.Elevator.Id] = e
			}
			switch msgFromNet.Event {
			case NewHall:
				if msgFromNet.Elevator.Id != thisElev.Id {
					hall = addHallBTN(hall, msgFromNet.BtnEvent)
					updateHallLights(hall)
					fromReAssigner, err := reAssign(thisElev.Id, elevatorMap, hall)
					if err != nil {
						log.Print("None fatal error: \n", err)
					} else {
						sendToFsm(fromReAssigner)
					}
				}

			case NewCab:
				fromReAssigner, err := reAssign(thisElev.Id, elevatorMap, hall)
				if err != nil {
					log.Print("None fatal error: \n", err)
				} else {
					sendToFsm(fromReAssigner)
				}

			//case UpdateElevState:
			//	newElevatorState := msgFromNet.Elevator
			//	// Clears hall buttons if there are any hall btns to clare (redundancy)
			//	for f := 0; f < config.NumFloors; f++ {
			//		if (newElevatorState.Floor == f && hall[f][elevio.BT_HallUp] && newElevatorState.Dir == elevio.MD_Up) || (newElevatorState.Floor == f && newElevatorState.Dir == elevio.MD_Down && hall[f][elevio.BT_HallUp] && !newElevatorState.AnyCabOrdersAhead()) {
			//			hall = clareHallBTN(hall, elevio.ButtonEvent{f, elevio.BT_HallUp})
			//			updateHallLights(hall)
			//			chMsgToNetwork <- NetworkPackage{
			//				Event:    ClareHall,
			//				Elevator: newElevatorState,
			//				BtnEvent: elevio.ButtonEvent{f, elevio.BT_HallUp},
			//			}
			//		}
			//		if (newElevatorState.Floor == f && hall[f][elevio.BT_HallDown] && newElevatorState.Dir == elevio.MD_Down) || (newElevatorState.Floor == f && newElevatorState.Dir == elevio.MD_Up && hall[f][elevio.BT_HallDown] && !newElevatorState.AnyCabOrdersAhead()) {
			//			hall = clareHallBTN(hall, elevio.ButtonEvent{f, elevio.BT_HallDown})
			//			updateHallLights(hall)
			//			chMsgToNetwork <- NetworkPackage{
			//				Event:    ClareHall,
			//				Elevator: newElevatorState,
			//				BtnEvent: elevio.ButtonEvent{f, elevio.MD_Down},
			//			}
			//			break
			//		}
			//	}

			case ClareHall:
				if msgFromNet.Elevator.Id != thisElev.Id {
					hall = clareHallBTN(hall, msgFromNet.BtnEvent)
					updateHallLights(hall)
				} else {
					log.Print("Hall button clared by self")
				}
				//hall = clareHallBTN(hall, msgFromNet.BtnEvent)
				//updateHallLights(hall)

			case RecoveredElevator:
				e := elevatorMap[msgFromNet.Elevator.Id]
				e.Elevator = msgFromNet.Elevator
				e.Alive = true
				elevatorMap[msgFromNet.Elevator.Id] = e
			}

		//case <-reRunTimer.C:
		//	reRunTimer.Reset(reRunRate)
		//	fromReAssigner, err := reAssign(thisElev.Id, elevatorMap, hall)
		//	if err != nil {
		//		log.Print("None fatal error: \n", err)
		//	} else {
		//		sendToFsm(fromReAssigner)
		//	}

		case p := <-chPeerUpdate:
			for _, id := range p.Lost {
				if e, ok := elevatorMap[id]; ok && id != thisElev.Id {
					// PeerUpdate sometimes registers that itself is lost without actually beeing lost
					// this fixes that problem
					e.Alive = false
					elevatorMap[id] = e
					fmt.Printf("We lost %s\n", id)
				}
			}

			if e, ok := elevatorMap[p.New]; ok && !e.Alive {
				if e.Elevator.Id == thisElev.Id {
					log.Printf("Witnesed my own death")
				}
				fmt.Printf("Would be cool if %s resurected \n", p.New)
				fmt.Printf(e.Elevator.String())
				fmt.Println()
			}
			//default:
			//	continue
		}
	}
}
