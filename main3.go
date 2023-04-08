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

const (
	udpPeer = 6001
	udpData = 6201
)

func main() {
	var id string
	var port string

	flag.StringVar(&port, "port", "", "Port of this elevator")
	flag.StringVar(&id, "id", "", "id of this elevator")
	flag.Parse()

	// Channels for networkMessaging
	chIoButtons := make(chan elevio.ButtonEvent, 100)
	chMsgToNetwork := make(chan messageHandler.NetworkPackage)
	chMsgFromNetwork := make(chan messageHandler.NetworkPackage)

	chPeerUpdate := make(chan peers.PeerUpdate)
	chPeerTxEnable := make(chan bool)

	chNewState := make(chan elevator.Elevator)

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

	// Goroutine for networkMessaging
	go peers.Transmitter(udpPeer, id, chPeerTxEnable)
	go peers.Receiver(udpPeer, chPeerUpdate)

	go bcast.Transmitter(udpData, chMsgToNetwork)
	go bcast.Receiver(udpData, chMsgFromNetwork)

	eObj := boot.Elevator(id, port, chIoFloor)

	go FSM2.FSM2(
		eObj,
		chIoFloor,
		chIoObstical,
		chIoStop,
		chNewState,
		chAddButton,
		chRmButton)

	go messageHandler.Handel(
		eObj,
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
