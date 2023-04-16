package messageHandler

import (
	"Project/config"
	"Project/elevio"
	"Project/localElevator/elevator"
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
	chFromFsm <-chan elevator.Elevator,
	chAddButtonToFsm chan<- elevio.ButtonEvent,
	chRmButtonFromFsm chan<- elevio.ButtonEvent,
	chPeerUpdate <-chan peers.PeerUpdate,
) {

	thisElev := elevator
	elevatorMap := make(map[string]ElevatorUpdate)
	elevatorMap[thisElev.Id] = ElevatorUpdate{*thisElev, true, 0}
	hall := make([][2]bool, config.NumFloors)
	//reRunRate := 200 * time.Millisecond
	//reRunTimer := time.NewTimer(reRunRate)
	LocalHall := make([][2]bool, config.NumFloors)
	for i:=0; i<config.NumFloors; i++ {
		LocalHall[i][0] = false
		LocalHall[i][1] = false
	}

	// Anonymous function that handles the sending to the fsm
	sendToFsm := func(fromReAssigner []assignValue) {
		for _, val := range fromReAssigner {
			time.Sleep(config.PollRate)
			if val.Type == Add {
				fmt.Print("REASSIgn ADDER \n")
				chAddButtonToFsm <- val.BtnEvent
				//LocalHall[val.BtnEvent.Floor][val.BtnEvent.Button] = true
			} else if val.Type == Remove {
				fmt.Print("Reassign fjerner \n")
				chRmButtonFromFsm <- val.BtnEvent
				LocalHall[val.BtnEvent.Floor][val.BtnEvent.Button] = false
			}
		}
	}
	chMsgToNetwork <- NetworkPackage{
		Event:    UpdateElevState,
		Elevator: *thisElev,
	}
	printHandlerStates := true
	for {
		select {
		case ioButton := <-chIoButtons:
			if printHandlerStates {fmt.Print("Message Handler: New Button \n")}
			if ioButton.Button == elevio.BT_Cab {
				e := elevatorMap[thisElev.Id]
				e.Elevator.AddOrder(ioButton)
				elevatorMap[thisElev.Id] = e
				fromReAssigner, err := reAssign(thisElev.Id, elevatorMap, hall)
				if err != nil {
					log.Print("None fatal error: \n", err)
				} else {
					sendToFsm(fromReAssigner)
				}
				chAddButtonToFsm <- ioButton
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

		case newElevatorState := <-chFromFsm:
			if printHandlerStates {fmt.Print("Message Handler: New ElevState \n")}
			e := elevatorMap[thisElev.Id]
			e.Elevator = newElevatorState
			elevatorMap[thisElev.Id] = e
			NewLocalHall := newElevatorState.GetHallOrders()
			// Clears hall buttons
			fmt.Print("THE GLOABL HALLS ")
			fmt.Print(hall,"\n")
			fmt.Print("LOCAL HALLS \n")
			fmt.Print(LocalHall,"\n")
			fmt.Print("NEW LOCAL HALLS \n")
			fmt.Print(NewLocalHall,"\n")
			for f := 0; f < config.NumFloors; f++ {
				for b := elevio.ButtonType(0); b < 2; b++ {
					if LocalHall[f][b] && !NewLocalHall[f][b] {//&& newElevatorState.Floor == f {
						/* fmt.Print(" YA YA VI CLEARE ORDRE \n") */
						hall = clareHallBTN(hall, elevio.ButtonEvent{f, b})
						chMsgToNetwork <- NetworkPackage{
							Event:    ClareHall,
							Elevator: newElevatorState,
							BtnEvent: elevio.ButtonEvent{f, b},
						}
						updateHallLights(hall)
					}
				}
			}
			LocalHall = NewLocalHall
			chMsgToNetwork <- NetworkPackage{
				Event:    UpdateElevState,
				Elevator: e.Elevator,
			}

		case msgFromNet := <-chMsgFromNetwork:
			//reRunTimer.Reset(reRunRate)
			//oldNet := elevatorMap[msgFromNet.Elevator.Id]
			if printHandlerStates {fmt.Print("Message Handler: From Network \n")}
			// Denne sjekker om det er en recover message fra nettet.
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
			// Denne sjekker om man må legge til den sendte heisen i elevatormappet.
			} else if _, ok := elevatorMap[msgFromNet.Elevator.Id]; !ok {
				// Have not seen this elevator before
				newElevator := ElevatorUpdate{msgFromNet.Elevator, true, 0}
				elevatorMap[msgFromNet.Elevator.Id] = newElevator
				this := elevatorMap[thisElev.Id] // Brodcast to net so the new elevator see the first elevator
				chMsgToNetwork <- NetworkPackage{
					Event:    UpdateElevState,
					Elevator: this.Elevator,
				}
			// Denne forsøker å revive.
			} else if e, ok := elevatorMap[msgFromNet.Elevator.Id]; ok && !e.Alive {
				// Her forsøker jeg å revive heisen når den først er registrert
				elevatorMap[e.Elevator.Id] = e
				e.Alive = true
				elevatorMap[e.Elevator.Id] = e
				fmt.Printf("Gammel heis sett på nett, sender states: %s\n", e.Elevator.Id)

				chMsgToNetwork <- NetworkPackage{
					Event:    Recover,
					Elevator: e.Elevator,
				}
				break
			// Denne bare oppdaterer elevatorMapet, med den nye versjonen av heisen.
			} else {
				e := elevatorMap[msgFromNet.Elevator.Id]
				e.Elevator = msgFromNet.Elevator
				e.Version++
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
			// ER DENNE NØDVEDNIG.
			case NewCab:
				fromReAssigner, err := reAssign(thisElev.Id, elevatorMap, hall)
				if err != nil {
					log.Print("None fatal error: \n", err)
				} else {
					sendToFsm(fromReAssigner)
				}
			// Får inn clear hall fra andre heiser, og clearer det. 
			case ClareHall:
				hall = clareHallBTN(hall, msgFromNet.BtnEvent)
				updateHallLights(hall)
			
			// Ekke så sikker. THomasito
			case RecoveredElevator:
				e := elevatorMap[msgFromNet.Elevator.Id]
				e.Elevator = msgFromNet.Elevator
				e.Alive = true
				e.Version++
				elevatorMap[msgFromNet.Elevator.Id] = e

/* 			case UpdateElevState:
				oldNetLocalHall := oldNet.Elevator.GetHallOrders()
				newNetLocalHall := msgFromNet.Elevator.GetHallOrders()


				//NewLocalHall := newElevatorState.GetHallOrders()
				// Clears hall buttons
				for f := 0; f < config.NumFloors; f++ {
					for b := elevio.ButtonType(0); b < 2; b++ {
						if oldNetLocalHall[f][b] && !newNetLocalHall[f][b] && msgFromNet.Elevator.Floor == f {
							hall = clareHallBTN(hall, elevio.ButtonEvent{f, b})
							chMsgToNetwork <- NetworkPackage{
								Event:    ClareHall,
								Elevator: msgFromNet.Elevator,
								BtnEvent: elevio.ButtonEvent{f, b},
							}
							updateHallLights(hall)
						}
					}
				} */

				
				

			default:
				continue
			}
		

		/* case <-reRunTimer.C:
			fmt.Print(hall)
			reRunTimer.Reset(reRunRate)
			fmt.Print("\n")
 */
/* 		case <-reRunTimer.C:
			fmt.Print(" Re Run Timer \n")
			reRunTimer.Reset(reRunRate)
			fromReAssigner, err := reAssign(thisElev.Id, elevatorMap, hall)
			if err != nil {
				log.Print("None fatal error: \n", err)
			} else {
				sendToFsm(fromReAssigner)
			} */

		case p := <-chPeerUpdate:
			if printHandlerStates {fmt.Print("Message Handler: Peer Update \n")}
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
		default:
			continue
		}
	}
}
