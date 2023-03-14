package distributor

import (
	"Project/network/peers"
	"Project/singleElevator/elevator"
	"Project/singleElevator/elevio"
	"fmt"
	"time"
)

type distElevator struct {
	Elevator elevator.Elevator
	Alive    bool
}

// ???Når vi deler states. Burde jeg dele alt jeg vet om alle andre
// elle holder det å dele meg selv

func Distribute(
	pid string,
	chButtons chan<- elevio.ButtonEvent,
	chMessageFromNetwork <-chan elevator.Elevator,
	chPeerUpdate <-chan peers.PeerUpdate,
	chRecovElevToNet chan<- elevator.Elevator,
	chRecovElevFromNet <-chan elevator.Elevator) {

	// Init elevator map and copy of elevator states
	elevatorMap := make(map[string]distElevator)
	var thisElevatorCopy elevator.Elevator

	for {
		select {
		case elevObj := <-chMessageFromNetwork:
			// Hvis chMessage er heisen selv
			if elevObj.Id == pid {
				thisElevatorCopy = elevObj
				_ = thisElevatorCopy // Supress error message
			}
			// Hvis ikke sett denne heisen før
			if _, ok := elevatorMap[elevObj.Id]; !ok {
				newElevator := distElevator{elevObj, true}
				elevatorMap[elevObj.Id] = newElevator
			} else {
				elevatorMap[elevObj.Id] = distElevator{elevObj, true} // Oppdterer heisen med nye states
				if elevObj.Id != pid {
					fmt.Printf("Recived updated states from %s\n", elevObj.Id)
				}
			}

		case p := <-chPeerUpdate:
			fmt.Printf("Peer uptade: \n")
			fmt.Printf("Peers %q\n", p.Peers)
			fmt.Printf(" New: %q\n", p.New)
			fmt.Printf(" Lost: %q\n", p.Lost)

			for _, val := range p.Lost {
				if e, ok := elevatorMap[val]; ok {
					e.Alive = false
					elevatorMap[val] = e
				}
			}

			// If id exist in map and is sett to not alive
			if e, ok := elevatorMap[p.New]; ok && !e.Alive {
				if e.Elevator.Id == pid {
					fmt.Printf("Jeg så meg selv dø??\n")
				}
				e.Alive = true
				elevatorMap[e.Elevator.Id] = e
				chRecovElevToNet <- e.Elevator
			}

		case recoverElev := <-chRecovElevFromNet:
			if recoverElev.Id == pid {
				fmt.Printf("Forsøk og recover\n")
				e := recoverElev
				for f := range e.Orders {
					for btn := range e.Orders[f] {
						fmt.Printf("Looped \n")
						if e.Orders[f][btn] {
							fmt.Printf("Sender gamle states \n")
							chButtons <- elevio.ButtonEvent{
								Floor:  f,
								Button: elevio.ButtonType(int(btn))}
							time.Sleep(50 * time.Millisecond)
						}
					}
				}
			}
		}
	}

}
