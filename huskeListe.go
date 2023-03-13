// Program som kjøres i egen terminal og laget for å huske statsene til
// elevator program tråder
package main

import (
	"Project/network/bcast"
	"Project/network/peers"
	elevator "Project/singleElevator/elevator"
	"fmt"
	"reflect"
)

func isPointer(value interface{}) bool {
	return reflect.ValueOf(value).Kind() == reflect.Ptr
}

type updateElevator struct {
	Elevator elevator.Elevator
	Alive    bool
}

func main() {
	var udpPeer = 6000
	var updData = 6200
	var udpRecover = 6400

	elevators := make(map[string]updateElevator)

	var elevCopy elevator.Elevator

	chMsgToNetwork := make(chan elevator.Elevator)
	chMsgFromNetwork := make(chan elevator.Elevator)

	chPeerUpdate := make(chan peers.PeerUpdate)
	chPeerTxEnable := make(chan bool)

	chRecovElevToNet := make(chan elevator.Elevator)
	chRecovElevFromNet := make(chan elevator.Elevator)
	go bcast.Transmitter(udpRecover, chRecovElevToNet)
	go bcast.Receiver(udpRecover, chRecovElevFromNet)

	// ** Network **
	var id = "420"
	// Sending data on the network
	go bcast.Transmitter(updData, chMsgToNetwork)
	go bcast.Receiver(updData, chMsgFromNetwork)

	// Keep track of who is alive on the network
	go peers.Transmitter(udpPeer, id, chPeerTxEnable)
	go peers.Receiver(udpPeer, chPeerUpdate)

	for {
		select {
		case elevObj := <-chMsgFromNetwork:
			elevCopy = elevObj
			if _, ok := elevators[elevCopy.Id]; !ok {
				newElevator := updateElevator{elevCopy, true}
				elevators[elevCopy.Id] = newElevator
				fmt.Printf("Added new elevator with id %s \n", elevCopy.Id)
			} else if e, ok := elevators[elevCopy.Id]; ok && !e.Alive {
				chRecovElevToNet <- e.Elevator
				e.Alive = true
				fmt.Printf("Elevator resurected %s, states sent \n", e.Elevator.Id)
				elevators[e.Elevator.Id] = e
			} else {
				elevators[elevCopy.Id] = updateElevator{elevCopy, true}
				fmt.Printf("Recived updated states \n")
			}

		case p := <-chPeerUpdate:
			fmt.Printf("Peer uptade: \n")
			fmt.Printf("Peers %q\n", p.Peers)
			fmt.Printf(" New: %q\n", p.New)
			fmt.Printf(" Lost: %q\n", p.Lost)

			for _, val := range p.Lost {
				if e, ok := elevators[val]; ok {
					e.Alive = false
					elevators[val] = e
				}
			}

			//if e, ok := elevators[p.New]; ok && !e.Alive {
			//	chRecovElevToNet <- e.Elevator
			//	e.Alive = true
			//	fmt.Printf("Elevator resurected [Ping] %s, states sent \n", e.Elevator.Id)
			//	elevators[e.Elevator.Id] = e
			//}

		}
	}

}
