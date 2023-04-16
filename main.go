package main

import (
	"Project/config"
	"Project/elevio"
	"Project/localElevator/boot"
	"Project/localElevator/fsm"
	"Project/messageHandler"
	"Project/network/T_SR"
	"Project/network/bcast"
	"Project/network/peers"
	"flag"
	"fmt"
)

func main() {
	var id string
	var port string
	udpPeer := 6001
	udpData := 6002
	var numFloors int
	var guiOn bool
	var soundOn bool

	flag.StringVar(&port, "port", "", "Port of this elevator")
	flag.StringVar(&id, "id", "", "id of this elevator")
	flag.IntVar(&numFloors, "floors", 4, "number of elevator floors")
	flag.BoolVar(&guiOn, "gui", false, "turn on gui")
	flag.BoolVar(&soundOn, "sound", false, "turn on sound")
	flag.Parse()
	config.NumFloors = numFloors // Update config
	fmt.Printf("Sound on%v\n", soundOn)

	// Channels for networkMessaging
	chIoButtons := make(chan elevio.ButtonEvent, 1)
	chMsgToNetwork := make(chan T_SR.SROnNet, 100) // Buffer sånn at får med alle button press selv om kanskje dæver(?)
	chMsgFromNetwork := make(chan T_SR.SROnNet, 100)

	chMsgToSend := make(chan messageHandler.NetworkPackage, 100)
	chMsgFromReciver := make(chan messageHandler.NetworkPackage, 100)

	chPeerUpdate := make(chan peers.PeerUpdate)
	chPeerTxEnable := make(chan bool)

	// Buffer so fsm does not need to wait for messageHandler
	chNewState := make(chan fsm.FsmOutput, 1000) // State updates from fsm to messageHandler

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
	eObj := boot.Elevator(id, port, chIoFloor, numFloors)
	eObjCopy := eObj.Clone()
	// Goroutine for networkMessaging
	go peers.Transmitter(udpPeer, id, chPeerTxEnable)
	go peers.Receiver(udpPeer, chPeerUpdate)

	go bcast.Transmitter(udpData, chMsgToNetwork)
	go bcast.Receiver(udpData, chMsgFromNetwork)

	go T_SR.Send(
		chMsgToNetwork,
		chMsgToSend,
	)

	go T_SR.Recive(
		chMsgFromNetwork,
		chMsgFromReciver)

	go fsm.FsmTest(
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
		chMsgFromReciver,
		chMsgToSend,
		chNewState,
		chAddButton,
		chRmButton,
		chPeerUpdate,
	)
	select {}

}
