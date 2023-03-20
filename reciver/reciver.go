package reciver

import (
	"Project/assigner"
	"Project/config"
	"Project/elevio"
	"Project/localElevator/elevator"
	"Project/network/peers"
	"fmt"
)

type reciveElevator struct {
	Elevator elevator.Elevator
	Alive    bool
	version  int
}

func Run(
	elevator elevator.Elevator,
	chIoButtons <-chan elevio.ButtonEvent,
	chIoFloor <-chan int,
	chIoObstical <-chan bool,
	chIoStop <-chan bool,
	chMsgFromNetwork <-chan elevator.Elevator,
	chReAssign chan<- map[string][][3]bool,
	chVirtualFloor chan<- int,
	chFromFSM <-chan elevator.Elevator,
	chMsgToNetwork chan<- elevator.Elevator,
	chPeerUpdate <-chan peers.PeerUpdate) {

	// Init variables
	thisElev := elevator
	elevatorMap := make(map[string]reciveElevator)
	elevatorMap[thisElev.Id] = reciveElevator{thisElev, true, 0}
	localhall := make([][2]bool, config.NumFloors)

	// Init this elevator i mai

	// Init disse

	for {
		select {
		case ioBtn := <-chIoButtons:
			if ioBtn.Button == elevio.BT_Cab {
				r := elevatorMap[thisElev.Id]
				r.Elevator.AddOrder(ioBtn)
				//r.version++
				elevatorMap[thisElev.Id] = r
			} else {
				r := elevatorMap[thisElev.Id]
				r.Elevator.AddOrder(ioBtn)
				r.version++
				elevatorMap[thisElev.Id] = r
				localhall = addTrue(localhall, r.Elevator.ToHallReq())
			}
			go reRunCost(elevatorMap, chReAssign, chMsgToNetwork, thisElev)

		case updateThis := <-chFromFSM:
			r := elevatorMap[thisElev.Id]
			r.Elevator = updateThis
			r.version++
			elevatorMap[thisElev.Id] = r
			localhall = addFalse(localhall, r.Elevator.ToHallReq())
			chMsgToNetwork <- r.Elevator
			go reRunCost(elevatorMap, chReAssign, chMsgToNetwork, thisElev)

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
			// Elevator is reborn
			if e, ok := elevatorMap[p.New]; ok && !e.Alive {
				if e.Elevator.Id == thisElev.Id {
					fmt.Printf("Jeg så meg selv dø??\n")
				}
				e.Alive = true
				elevatorMap[e.Elevator.Id] = e
				//chRecovElevToNet <- e.Elevator // Possible lock
			}
			go reRunCost(elevatorMap, chReAssign, chMsgToNetwork, thisElev)

		case elevObj := <-chMsgFromNetwork:
			if elevObj.Id != thisElev.Id {
				continue
				// If have not seen this elevator before
			} else if _, ok := elevatorMap[elevObj.Id]; !ok {
				fmt.Printf("New elevator %s\n", elevObj.Id)
				newElevator := reciveElevator{elevObj, true, 0}
				elevatorMap[elevObj.Id] = newElevator
				go reRunCost(elevatorMap, chReAssign, chMsgToNetwork, thisElev)

			} else {
				oldElevator := elevatorMap[elevObj.Id]
				oldElevator.Elevator = elevObj
				oldElevator.version++
				elevatorMap[elevObj.Id] = oldElevator
				localhall = addTrue(localhall, oldElevator.Elevator.ToHallReq())
				go reRunCost(elevatorMap, chReAssign, chMsgToNetwork, thisElev)
			}
		}
	}
}

func reRunCost(elevatorMap map[string]reciveElevator,
	chReAssign chan<- map[string][][3]bool,
	chMsgToNetwork chan<- elevator.Elevator,
	thisElev elevator.Elevator) {
	chMsgToNetwork <- elevatorMap[thisElev.Id].Elevator
	input := config.HRAInput{
		States:       make(map[string]config.HRAElevState),
		HallRequests: make([][2]bool, config.NumFloors),
	}
	for id, val := range elevatorMap {
		if val.Alive {
			fmt.Printf("Elevator %s is alive\n", id)
			hraElev := val.Elevator.ToHRA()
			input.States[id] = hraElev
		}
	}
	chReAssign <- assigner.Assign(input)
}

// addTrue - changes value to true if true in addfrom
func addTrue(addTO, addFrom [][2]bool) [][2]bool {

	for i := range addFrom {
		if addFrom[i][0] {
			addTO[i][0] = true
		}
		if addFrom[i][1] {
			addTO[i][1] = true
		}
	}
	return addTO
}

func addFalse(addTO, addFrom [][2]bool) [][2]bool {
	for i := range addFrom {
		if !addFrom[i][0] {
			addTO[i][0] = false
		}
		if !addFrom[i][1] {
			addTO[i][1] = false
		}
	}
	return addTO
}
