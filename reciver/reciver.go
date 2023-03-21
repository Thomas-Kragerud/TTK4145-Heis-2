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
	chPeerUpdate <-chan peers.PeerUpdate,
	chAddBtnNet chan elevio.ButtonEvent,
	chRmBtnNet chan elevio.ButtonEvent,
	chClareHallFsm chan elevio.ButtonEvent,
	chReciveBtnNet chan elevio.ButtonEvent,
	chRmReciveBtnNet chan elevio.ButtonEvent) {

	// Init variables
	thisElev := elevator
	elevatorMap := make(map[string]reciveElevator)
	elevatorMap[thisElev.Id] = reciveElevator{thisElev, true, 0}
	hall := make([][2]bool, config.NumFloors)

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
				chMsgToNetwork <- r.Elevator
			} else {
				r := elevatorMap[thisElev.Id]
				r.Elevator.AddOrder(ioBtn)
				r.version++
				elevatorMap[thisElev.Id] = r
				chAddBtnNet <- ioBtn
				//localhall = addTrue(localhall, r.Elevator.ToHallReq())
			}

			//chMsgToNetwork <- elevatorMap[thisElev.Id].Elevator

		case ioBtnNet := <-chReciveBtnNet:
			hall = addBTN(hall, ioBtnNet)
			go reRunCost(elevatorMap, chReAssign, hall)

		case c := <-chClareHallFsm:
			chRmBtnNet <- c

		case rmBtnNet := <-chRmReciveBtnNet:
			hall = rmBTN(hall, rmBtnNet)
			go reRunCost(elevatorMap, chReAssign, hall)

		case updateThis := <-chFromFSM:
			r := elevatorMap[thisElev.Id]
			r.Elevator = updateThis
			r.version++
			elevatorMap[thisElev.Id] = r
			chMsgToNetwork <- r.Elevator
			//go reRunCost(elevatorMap, chReAssign, hall)

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
			go reRunCost(elevatorMap, chReAssign, hall)

		case elevObj := <-chMsgFromNetwork:
			if elevObj.Id == thisElev.Id {
				r := elevatorMap[thisElev.Id]
				r.Elevator = elevObj
				r.version++
				elevatorMap[thisElev.Id] = r
				go reRunCost(elevatorMap, chReAssign, hall)
				// If have not seen this elevator before
			} else if _, ok := elevatorMap[elevObj.Id]; !ok {
				fmt.Printf("New elevator %s\n", elevObj.Id)
				newElevator := reciveElevator{elevObj, true, 0}
				elevatorMap[elevObj.Id] = newElevator
				go reRunCost(elevatorMap, chReAssign, hall)
			} else {
				oldElevator := elevatorMap[elevObj.Id]
				oldElevator.Elevator = elevObj
				oldElevator.version++
				elevatorMap[elevObj.Id] = oldElevator

				//localhall = addTrue(localhall, oldElevator.Elevator.ToHallReq())
				go reRunCost(elevatorMap, chReAssign, hall)
			}
		default:
			continue
		}
	}
}

func reRunCost(elevatorMap map[string]reciveElevator,
	chReAssign chan<- map[string][][3]bool,
	hall [][2]bool) {
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
	input.HallRequests = hall
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

func addBTN(hall [][2]bool, btn elevio.ButtonEvent) [][2]bool {
	if btn.Button == elevio.BT_HallUp {
		hall[btn.Floor][0] = true
	} else if btn.Button == elevio.BT_HallDown {
		hall[btn.Floor][1] = true
	}
	return hall
}

func rmBTN(hall [][2]bool, btn elevio.ButtonEvent) [][2]bool {
	if btn.Button == elevio.BT_HallUp {
		hall[btn.Floor][0] = false
	} else if btn.Button == elevio.BT_HallDown {
		hall[btn.Floor][1] = false
	}
	return hall
}
