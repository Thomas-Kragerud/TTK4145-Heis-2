package reciver

import (
	"Project/assigner"
	"Project/config"
	"Project/elevio"
	"Project/localElevator/elevator"
	"fmt"
)

var _stateToString = map[elevator.ElevatorState]string{
	elevator.Idle:     "idle",
	elevator.DoorOpen: "doorOpen",
	elevator.Moving:   "moving",
}


type ElevNetworkMessage struct {
	id string
	floor int
	thisElevatorState elevator.ElevatorState
	thisElevatorDir elevio.MotorDirection
	thisElevatorCabRequests []bool
	ElevatorHallRequests [][]config.OrderState
}

type LocalElevator struct {
	id string
	floor int
	thisElevatorState elevator.ElevatorState
	thisElevatorDir elevio.MotorDirection
	thisElevatorCabRequests []bool
}

func printLocalElevatorData(LocalElevatorData []*LocalElevator) {
	for _, elev := range LocalElevatorData {
		str := "***Elevator***\n"
		str += fmt.Sprintf("Floor: %d\n", elev.floor)
		str += fmt.Sprintf("Direction: %s\n", elevio.ToStringMotorDirection(elev.thisElevatorDir))
		str += fmt.Sprintf("State: %s\n", _stateToString[elev.thisElevatorState])
		str += "Orders:\n"
		str += " cab\n"
		for floor := range elev.thisElevatorCabRequests {
			str += fmt.Sprintf("| %d", floor)
			str += fmt.Sprintf(" %v ", elev.thisElevatorCabRequests[floor])
			str += "|\n"
		}
		fmt.Printf(str)
	}
}

func LocalElevatorInit(pid string) LocalElevator{
	cabRequests := make([]bool,4)

	for floor := 0; floor < config.NumFloors; floor++{
		cabRequests[floor] = true
	}
	return LocalElevator{id: "thomas", floor: 1, thisElevatorState: elevator.Idle, thisElevatorDir: elevio.MD_Down, thisElevatorCabRequests: cabRequests}
	//return LocalElevator(id: pid, floor: 0, thisElevatorState: Idle, thisElevatorDir: MD_Stop, thisElevatorCabRequests: cabRequests)
}

func updateeLocalElevatorData(localElevatorData []*LocalElevator, newElevator ElevNetworkMessage) {
	for _,localElev := range localElevatorData {
		if localElev.id == newElevator.id {
			localElev.floor = newElevator.floor
			localElev.thisElevatorState = newElevator.thisElevatorState
			localElev.thisElevatorDir = newElevator.thisElevatorDir
			localElev.thisElevatorCabRequests = newElevator.thisElevatorCabRequests
			return
		}
	}
	addElevatorToLocalElevatorData(&localElevatorData, newElevator)
}


func addElevatorToLocalElevatorData(localElevatorData *[]*LocalElevator, newElevator ElevNetworkMessage) {
	tempElev := new(LocalElevator)
	(*tempElev).floor = newElevator.floor
	(*tempElev).thisElevatorState = newElevator.thisElevatorState
	(*tempElev).thisElevatorDir = newElevator.thisElevatorDir
	(*tempElev).thisElevatorCabRequests = newElevator.thisElevatorCabRequests
	(*tempElev).id = newElevator.id
	*localElevatorData = append(*localElevatorData,tempElev)	
}

func broadcastLocalElevator(localElevatorData LocalElevator,localHallRequests [][]config.OrderState ,ch_ToNetwork chan<- ElevNetworkMessage) {
	tempElevatorMessage := new(ElevNetworkMessage)
	(*tempElevatorMessage).id = localElevatorData.id
	(*tempElevatorMessage).floor = localElevatorData.floor
	(*tempElevatorMessage).thisElevatorState = localElevatorData.thisElevatorState
	(*tempElevatorMessage).thisElevatorDir = localElevatorData.thisElevatorDir
	(*tempElevatorMessage).thisElevatorCabRequests = localElevatorData.thisElevatorCabRequests
	(*tempElevatorMessage).ElevatorHallRequests = localHallRequests
	ch_ToNetwork <- (*tempElevatorMessage)
}

func CommunicateWithNet(
	id string,
	chMsgFromNetwork <-chan ElevNetworkMessage, 
	chMsgToNetwork chan<- ElevNetworkMessage,
	chNewLocalOrder <-chan elevio.ButtonEvent,
	chNewLocalState <-chan elevator.Elevator,
	chToLocalElevator chan<- elevio.ButtonEvent,
	chFromLocalElevator <-chan elevator.Elevator) {

	
	localElevatorData := make([]*LocalElevator,0)
	thisElevator := new(LocalElevator)
	*thisElevator = LocalElevatorInit(id);
	localElevatorData = append(localElevatorData,thisElevator)
	//thisHallRequests := new([][]config.OrderState)
	//activeHallRequests := new([][]bool)



	for {
		switch {
			case newOrder := <- chNewLocalOrder:
				chToLocalElevator <- newOrder
			case upDateFromLocalElevator := <- chFromLocalElevator:
				for elev := range localElevatorData{
					if elev.id == upDateFromLocalElevator.id{
						elev.thisElevatorState = upDateFromLocalElevator.State
						elev.floor = upDateFromLocalElevator.Floor
						for floor := range config.numFloors{
							elev.thisElevatorCabRequests[floor][BT_Cab] = updateeLocalElevator.Orders[floor][BT_Cab]
						}
					}
				}
			case elevMessage := <- chMsgFromNetwork:
		
		}
	}
}




func TestingStuff(){
	elevator1 := new(LocalElevator)
	cabRequests := make([]bool,config.NumFloors)
	for floor := 0; floor < config.NumFloors; floor++{
		cabRequests[floor] = false
	}
	
	hallRequests := make([][]config.OrderState,4)
	for floor := 0; floor < config.NumFloors; floor++{
		hallRequests[floor] = make([]config.OrderState, 3)
	}
	*elevator1 = LocalElevator{id: "nils", floor: 2, thisElevatorState: elevator.Idle, thisElevatorDir: elevio.MD_Down, thisElevatorCabRequests: cabRequests}
	elevator2 := new(LocalElevator)
	*elevator2 = LocalElevatorInit("Thomas")
	localElevatorData2 := make([]*LocalElevator,0)
	localElevatorData2 = append(localElevatorData2,elevator1)
	//localElevatorData2 = addElevatorToLocalElevatorData(&localElevatorData2,elevator1)
	localElevatorData2 = append(localElevatorData2,elevator2)
	cabRequests[3] = true
	networkMessage := new(ElevNetworkMessage)
	*networkMessage = ElevNetworkMessage{id: "nils", floor: 4, thisElevatorState: elevator.Idle, thisElevatorDir: elevio.MD_Stop, thisElevatorCabRequests: cabRequests, ElevatorHallRequests: hallRequests}
	updateeLocalElevatorData(localElevatorData2,*networkMessage)
	printLocalElevatorData(localElevatorData2)
}




func Run(
	pid string,
	chIoButtons <-chan elevio.ButtonEvent,
	chIoFloor <-chan int,
	chIoObstical <-chan bool,
	chIoStop <-chan bool) {

	// Init this elevator i main
	var thisElev config.SendElev
	allStates := make(map[string]config.SendElev)
	var hR config.SendHall

	var data config.InputData
	output := make(map[string][][]bool)

	thisElev.Init()
	hR.Init()
	allStates[pid] = thisElev

	data.HallRequests.HallRequests = hR.HallRequests
	data.States = allStates

	// Init disse

	for {
		select {
		case ioBtn := <-chIoButtons:
			if ioBtn.Button == elevio.BT_Cab {
				allStates[pid].CabRequests[ioBtn.Floor] = true
				data.HallRequests = hR
				// Formater til input data objekt og send til assigner

				fmt.Printf("Recived cat call at %d\n", ioBtn.Floor)
			} else {
				hR.Update(ioBtn.Floor, ioBtn.Button)
				fmt.Printf("Recived hallbtn at %d", ioBtn.Floor)
				// Noe om at den ikke er assigned
			}
			data.HallRequests.HallRequests = hR.HallRequests
			data.States = allStates
			output = assigner.Assign(data)
			fmt.Printf("Output: %v\n", output)

			break // Trenger jeg ???

		case ioF := <-chIoFloor:
			thisElev.Floor = ioF
			allStates[pid] = thisElev
			fmt.Printf("Recived new floor %d", ioF)

			//case ioObst := <-chIoObstical:
			//
			//
			//case ioS := <-chIoStop:
			//
			//case card := <-chBrodcastet:

		}
	}
}