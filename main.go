package main

import (
	"Project/config"
	"Project/elevio"
	"Project/localElevator/boot"
	"Project/localElevator/elevator"
	"Project/localElevator/fsm"
	"Project/messageHandler"
	"Project/network/networkBridge"
	"Project/network/bcast"
	"Project/network/peers"
	"flag"
)

func main() {
	var id string
	var port string
	udpPeer := 6001
	udpData := 6002
	var numFloors int


	// Parse flags
	flag.StringVar(&port, "port", "", "Port of this elevator")
	flag.StringVar(&id, "id", "", "id of this elevator")
	flag.IntVar(&numFloors, "floors", 4, "number of elevator floors")
	flag.Parse()

	// Set number of floors
	config.NumFloors = numFloors // Update config

	// Channels for networkMessaging
	chIoButtons := make(chan elevio.ButtonEvent, 1)
	chMsgToNetwork := make(chan networkBridge.SROnNet, 100) // Buffer sånn at får med alle button press selv om kanskje dæver(?)
	chMsgFromNetwork := make(chan networkBridge.SROnNet, 100)

	// Channels for networkBridge
	chMsgToNBSend := make(chan messageHandler.NetworkPackage, 100)
	chMsgFromNBReciver := make(chan messageHandler.NetworkPackage, 100)

	// Channels for network package
	chPeerUpdate := make(chan peers.PeerUpdate)
	chPeerTxEnable := make(chan bool)

	// Buffer so fsm does not need to wait for messageHandler
	chNewState := make(chan elevator.Elevator, 100)

	// Channels for local elevator
	chIoFloor := make(chan int)
	chIoObstical := make(chan bool)
	chIoStop := make(chan bool)
	chAddButton := make(chan elevio.ButtonEvent, 1)
	chRmButton := make(chan elevio.ButtonEvent, 1)

	// Goroutines for interfacing with I/O
	go elevio.PollFloorSensor(chIoFloor)
	go elevio.PollObstructionSwitch(chIoObstical)
	go elevio.PollStopButton(chIoStop)
	go elevio.PollButtons(chIoButtons)

	// Boot elevator
	eObj := boot.Elevator(id, port, chIoFloor, numFloors)
	eObjCopy := eObj.Clone()
	
	// Goroutine for networkMessaging
	go peers.Transmitter(udpPeer, id, chPeerTxEnable)
	go peers.Receiver(udpPeer, chPeerUpdate)

	go bcast.Transmitter(udpData, chMsgToNetwork)
	go bcast.Receiver(udpData, chMsgFromNetwork)

	go networkBridge.Send(
		chMsgToNetwork,
		chMsgToNBSend,
	)

	go networkBridge.Recieve(
		chMsgFromNetwork,
		chMsgFromNBReciver, 
	)

	go fsm.Fsm(
		&eObj,
		chIoFloor,
		chIoObstical,
		chIoStop,
		chNewState,
		chRmButton,
		chAddButton,
	)

	go messageHandler.Handle(
		&eObjCopy,
		chIoButtons,
		chMsgFromNBReciver,
		chMsgToNBSend,
		chNewState,
		chAddButton,
		chRmButton,
		chPeerUpdate,
	)
	select {}

}
