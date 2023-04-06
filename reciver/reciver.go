package reciver

import (
	"Project/assigner"
	"Project/config"
	"Project/elevio"
	"Project/localElevator/elevator"
	"Project/network/peers"
	"Project/networkHandler"

	//"fmt"
	"sync"
)

type reciveElevator struct {
	Elevator elevator.Elevator
	Alive    bool
	Version  int
}

var mu sync.Mutex

func Run(
	elevator elevator.Elevator,
	chIoButtons <-chan elevio.ButtonEvent,
	chIoFloor <-chan int,
	chIoObstical <-chan bool,
	chIoStop <-chan bool,
	chMsgFromNetwork <-chan networkHandler.NetworkPackage,
	chReAssign chan<- map[string][][3]bool,
	chVirtualFloor chan<- int,
	chFromFSM <-chan elevator.Elevator,
	chMsgToNetwork chan<- networkHandler.NetworkPackage,
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
				elevatorMap[thisElev.Id] = r
				go reRunCost(elevatorMap, chReAssign, hall)
				msg := networkHandler.NetworkPackage{
					networkHandler.NewCab,
					r.Elevator,
					ioBtn}
				chMsgToNetwork <- msg

			} else {
				r := elevatorMap[thisElev.Id]
				r.Elevator.AddOrder(ioBtn)
				elevatorMap[thisElev.Id] = r
				//chAddBtnNet <- ioBtn
				msg := networkHandler.NetworkPackage{
					networkHandler.NewHall,
					r.Elevator,
					ioBtn}
				chMsgToNetwork <- msg
				// Case when internet is not working
			}

		case ioBtnNet := <-chReciveBtnNet:
			hall = addBTN(hall, ioBtnNet)
			updateHallLights(hall)
			go reRunCost(elevatorMap, chReAssign, hall)

		case c := <-chClareHallFsm:
			chRmBtnNet <- c

		case rmBtnNet := <-chRmReciveBtnNet:
			hall = rmBTN(hall, rmBtnNet)
			updateHallLights(hall)
			go reRunCost(elevatorMap, chReAssign, hall)

		case updateThis := <-chFromFSM:
			r := elevatorMap[thisElev.Id]
			r.Elevator = updateThis
			elevatorMap[thisElev.Id] = r
			msg := networkHandler.NetworkPackage{
				Event:    networkHandler.UpdateElevState,
				Elevator: r.Elevator}
			chMsgToNetwork <- msg
			//go reRunCost(elevatorMap, chReAssign, hall)

		case p := <-chPeerUpdate:
			//fmt.Printf("Peer uptade: \n")
			//fmt.Printf("Peers %q\n", p.Peers)
			//fmt.Printf(" New: %q\n", p.New)
			//fmt.Printf(" Lost: %q\n", p.Lost)
			for _, val := range p.Lost {
				if e, ok := elevatorMap[val]; ok {
					e.Alive = false
					elevatorMap[val] = e
				}
			}
			// Elevator is reborn
			if e, ok := elevatorMap[p.New]; ok && !e.Alive {
				if e.Elevator.Id == thisElev.Id {
					//fmt.Printf("Jeg så meg selv dø??\n")
				}
				e.Alive = true
				elevatorMap[e.Elevator.Id] = e
				//chRecovElevToNet <- e.Elevator // Possible lock
			}
			go reRunCost(elevatorMap, chReAssign, hall)

		case msgFromNet := <-chMsgFromNetwork:
			switch msgFromNet.Event {
			case networkHandler.NewCab:

			case networkHandler.UpdateElevState:

			case networkHandler.NewHall:
				// Just to see if works

				hall = addBTN(hall, msgFromNet.BtnEvent)
				updateHallLights(hall)
				go reRunCost(elevatorMap, chReAssign, hall)

			case networkHandler.AkHall:

			case networkHandler.RmHall:

			case networkHandler.AkRmHall:

			case networkHandler.PeriodicUpdate:

			}
			if msgFromNet.Elevator.Id == thisElev.Id {
				r := elevatorMap[thisElev.Id]
				r.Elevator = msgFromNet.Elevator
				r.Version++
				elevatorMap[thisElev.Id] = r
				// If have not seen this elevator before
			} else if _, ok := elevatorMap[msgFromNet.Elevator.Id]; !ok {
				//fmt.Printf("New elevator %s\n", elevObj.Id)
				newElevator := reciveElevator{msgFromNet.Elevator, true, 0}
				elevatorMap[msgFromNet.Elevator.Id] = newElevator
				this := elevatorMap[thisElev.Id] // Can be removed when by sending periodically
				msg := networkHandler.NetworkPackage{
					Event:    networkHandler.UpdateElevState,
					Elevator: this.Elevator}
				chMsgToNetwork <- msg

				go reRunCost(elevatorMap, chReAssign, hall)
			} else {
				oldElevator := elevatorMap[msgFromNet.Elevator.Id]
				oldElevator.Elevator = msgFromNet.Elevator
				oldElevator.Version++
				elevatorMap[msgFromNet.Elevator.Id] = oldElevator
				go reRunCost(elevatorMap, chReAssign, hall)
			}
			break

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
			//fmt.Printf("Elevator %s is alive\n", id)
			hraElev := val.Elevator.ToHRA()
			input.States[id] = hraElev
		}
	}
	input.HallRequests = hall
	chReAssign <- assigner.Assign(input)
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

func updateHallLights(hall [][2]bool) {
	for i := range hall {
		if hall[i][0] {
			elevio.SetButtonLamp(elevio.BT_HallUp, i, true)
		} else {
			elevio.SetButtonLamp(elevio.BT_HallUp, i, false)
		}
		if hall[i][1] {
			elevio.SetButtonLamp(elevio.BT_HallDown, i, true)
		} else {
			elevio.SetButtonLamp(elevio.BT_HallDown, i, false)
		}
	}
}
