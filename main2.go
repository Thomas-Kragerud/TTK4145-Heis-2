package main

import (
	"Project/elevio"
	"Project/localElevator/FSM2"
	"Project/localElevator/boot"
	"Project/localElevator/elevator"
	"Project/messageHandler"
	"Project/network/bcast"
	"Project/network/peers"
	"flag"
)

func main() {
	var id string
	var port string
	udpPeer := 6001
	udpData := 6002

	flag.StringVar(&port, "port", "", "Port of this elevator")
	flag.StringVar(&id, "id", "", "id of this elevator")
	flag.Parse()

	// Channels for networkMessaging
	chIoButtons := make(chan elevio.ButtonEvent, 100)
	chMsgToNetwork := make(chan messageHandler.NetworkPackage, 100) // Buffer sånn at får med alle button press selv om kanskje dæver(?)
	chMsgFromNetwork := make(chan messageHandler.NetworkPackage, 100)

	chPeerUpdate := make(chan peers.PeerUpdate)
	chPeerTxEnable := make(chan bool)

	// Buffer so fsm does not need to wait for messageHandler
	chNewState := make(chan elevator.Elevator, 100)

	// Channels for local elevator
	chIoFloor := make(chan int)
	chIoObstical := make(chan bool)
	chIoStop := make(chan bool)
	chAddButton := make(chan elevio.ButtonEvent, 100)
	chRmButton := make(chan elevio.ButtonEvent, 100)

	// Goroutines for interfacing with I/O
	go elevio.PollFloorSensor(chIoFloor)
	go elevio.PollObstructionSwitch(chIoObstical)
	go elevio.PollStopButton(chIoStop)
	go elevio.PollButtons(chIoButtons)

	// Boot elevator
	eObj := boot.Elevator(id, port, chIoFloor)
	eObjCopy := eObj.Clone()
	// Goroutine for networkMessaging
	go peers.Transmitter(udpPeer, id, chPeerTxEnable)
	go peers.Receiver(udpPeer, chPeerUpdate)

	go bcast.Transmitter(udpData, chMsgToNetwork)
	go bcast.Receiver(udpData, chMsgFromNetwork)

	go FSM2.FsmTest(
		&eObj,
		chIoFloor,
		chIoObstical,
		chIoStop,
		chNewState,
		chRmButton,
		chAddButton,
	)

	go messageHandler.Handel(
		&eObjCopy,
		chIoButtons,
		chMsgFromNetwork,
		chMsgToNetwork,
		chNewState,
		chAddButton,
		chRmButton,
		chPeerUpdate,
	)
	select {}

}
