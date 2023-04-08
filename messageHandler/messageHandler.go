package messageHandler

import (
	"Project/config"
	"Project/elevio"
	"Project/localElevator/elevator"
	"Project/network/peers"
	"fmt"
	"time"
)

func Handel(
	elevator elevator.Elevator,
	chIoButtons <-chan elevio.ButtonEvent,
	chMsgFromNetwork <-chan NetworkPackage,
	chMsgToNetwork chan<- NetworkPackage,
	chFromFsm <-chan elevator.Elevator,
	chAddButtonToFsm chan<- elevio.ButtonEvent,
	chRmButtonFromFsm chan<- elevio.ButtonEvent,
	chPeerUpdate <-chan peers.PeerUpdate,
) {

	thisElev := <-chFromFsm
	elevatorMap := make(map[string]ElevatorUpdate)
	elevatorMap[thisElev.Id] = ElevatorUpdate{thisElev, true, 0}
	hall := make([][2]bool, config.NumFloors)
	reRunRate := 2000 * time.Millisecond
	reRunTimer := time.NewTimer(reRunRate)

	// Anonumus function that handels the sending
	sendToFsm := func(fromReAssigner []assignValue) {
		for _, val := range fromReAssigner {
			if val.Type == Add {
				chAddButtonToFsm <- val.BtnEvent
			} else if val.Type == Remove {
				chRmButtonFromFsm <- val.BtnEvent
			}
		}

	}

	for {
		select {
		case ioButton := <-chIoButtons:
			if ioButton.Button == elevio.BT_Cab {
				e := elevatorMap[thisElev.Id]
				e.Elevator.AddOrder(ioButton)
				elevatorMap[thisElev.Id] = e
				chAddButtonToFsm <- ioButton
				fromReAssigner := reAssign(thisElev.Id, elevatorMap, hall)
				sendToFsm(fromReAssigner)
				chMsgToNetwork <- NetworkPackage{
					NewCab,
					e.Elevator,
					ioButton,
				}
			} else {
				e := elevatorMap[thisElev.Id]
				//e.Elevator.AddOrder(ioButton) // Only for bookkeeping
				hall = addHallBTN(hall, ioButton) // Add hall to this elevator list of hall
				updateHallLights(hall)
				msg := NetworkPackage{
					NewHall,
					e.Elevator,
					ioButton,
				}
				chMsgToNetwork <- msg
			}

		case newElevatorState := <-chFromFsm:
			e := elevatorMap[thisElev.Id]
			e.Elevator = newElevatorState
			elevatorMap[thisElev.Id] = e

			// Clears hall buttons
			for f := 0; f < config.NumFloors; f++ {
				for b := elevio.ButtonType(0); b < 2; b++ {
					if hall[f][b] && newElevatorState.Floor == f {
						hall = clareHallBTN(hall, elevio.ButtonEvent{f, b})
						updateHallLights(hall)
						chMsgToNetwork <- NetworkPackage{
							Event:    ClareHall,
							Elevator: newElevatorState,
							BtnEvent: elevio.ButtonEvent{f, b},
						}
						break
					}
				}
			}
			chMsgToNetwork <- NetworkPackage{
				Event:    UpdateElevState,
				Elevator: e.Elevator,
			}

		case msgFromNet := <-chMsgFromNetwork:
			if msgFromNet.Elevator.Id == thisElev.Id {
				if msgFromNet.Event == Recover {
					for f := 0; f < config.NumFloors; f++ {
						for b := elevio.ButtonType(0); b < 3; b++ {
							if msgFromNet.Elevator.Orders[f][b] {
								chAddButtonToFsm <- elevio.ButtonEvent{f, b}
							}
						}
					}
				}

			} else if _, ok := elevatorMap[msgFromNet.Elevator.Id]; !ok {
				// Have not seen this elevator before
				newElevator := ElevatorUpdate{msgFromNet.Elevator, true, 0}
				elevatorMap[msgFromNet.Elevator.Id] = newElevator
				this := elevatorMap[thisElev.Id] // Brodcast to net so the new elevator see the first elevator
				chMsgToNetwork <- NetworkPackage{
					Event:    UpdateElevState,
					Elevator: this.Elevator,
				}
			} else if e, ok := elevatorMap[msgFromNet.Elevator.Id]; ok && !e.Alive {
				// Her forsøker jeg å revive heisen når den først er registrert
				e.Alive = true
				elevatorMap[e.Elevator.Id] = e
				fmt.Printf("Trying to recover elevator but from another place %s", e.Elevator.Id)
				chMsgToNetwork <- NetworkPackage{
					Event:    Recover,
					Elevator: e.Elevator,
				}

			} else {
				e := elevatorMap[msgFromNet.Elevator.Id]
				e.Elevator = msgFromNet.Elevator
				e.Version++
				elevatorMap[msgFromNet.Elevator.Id] = e
			}
			switch msgFromNet.Event {
			case NewHall:
				hall = addHallBTN(hall, msgFromNet.BtnEvent)
				updateHallLights(hall)

				fromReAssigner := reAssign(thisElev.Id, elevatorMap, hall)
				sendToFsm(fromReAssigner)

			case NewCab:
				fromReAssigner := reAssign(thisElev.Id, elevatorMap, hall)
				sendToFsm(fromReAssigner)

			case UpdateElevState:
				newElevatorState := msgFromNet.Elevator
				// Clears hall buttons if there are any hall btns to clare (redundancy)
				for f := 0; f < config.NumFloors; f++ {
					for b := elevio.ButtonType(0); b < 2; b++ {
						if hall[f][b] && newElevatorState.Floor == f {
							hall = clareHallBTN(hall, elevio.ButtonEvent{f, b})
							updateHallLights(hall)
							chMsgToNetwork <- NetworkPackage{
								Event:    ClareHall,
								Elevator: newElevatorState,
								BtnEvent: elevio.ButtonEvent{f, b},
							}
							break
						}
					}
				}

			case ClareHall:
				hall = clareHallBTN(hall, msgFromNet.BtnEvent)
				updateHallLights(hall)

			}

		case <-reRunTimer.C:
			reRunTimer.Reset(reRunRate)
			fromReAssigner := reAssign(thisElev.Id, elevatorMap, hall)
			sendToFsm(fromReAssigner)

		case p := <-chPeerUpdate:
			for _, id := range p.Lost {
				if e, ok := elevatorMap[id]; ok {
					e.Alive = false
					elevatorMap[id] = e
					fmt.Printf("We lost %s\n", id)
				}
			}
			// Hva hvis den får peer update men ikke har noe kjennskap til de andre?
			// Elevator is reborn
			if e, ok := elevatorMap[p.New]; ok && !e.Alive {
				if e.Elevator.Id == thisElev.Id {
					//fmt.Printf("Jeg så meg selv dø??\n")
				}
				//e.Alive = true
				//elevatorMap[e.Elevator.Id] = e
				fmt.Printf("Would be cool if %s resurected \n", p.New)
				fmt.Printf(e.Elevator.String())
				fmt.Println()
				//chMsgToNetwork <- NetworkPackage{
				//	Event:    Recover,
				//	Elevator: e.Elevator,
				//}
			}
		}
	}

}
