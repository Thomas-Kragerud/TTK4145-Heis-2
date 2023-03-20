package main

import (
	"Project/distributor"
	"Project/elevio"
	"Project/localElevator/FSM"
	"Project/localElevator/boot"
	"Project/localElevator/elevator"
	"Project/network/bcast"
	"Project/network/peers"
	"Project/reciver"
	"flag"
)

func main() {
	// parameter setting
	var udpPeer = 6000
	var updData = 6200
	var udpRecover = 6400
	var id string
	var port string

	flag.StringVar(&port, "port", "", "Port of this elevator")
	flag.StringVar(&id, "id", "", "id of this elevator")
	flag.Parse()

	// ****** Set up channels ******

	// Channels for distribution
	chMsgToNetwork := make(chan elevator.Elevator)
	chMsgFromNetwork := make(chan elevator.Elevator)

	chRecovElevToNet := make(chan elevator.Elevator)
	chRecovElevFromNet := make(chan elevator.Elevator)

	chPeerUpdate := make(chan peers.PeerUpdate)
	chPeerTxEnable := make(chan bool)

	// Channels for local elevator
	chIoFloor := make(chan int)
	chIoObstical := make(chan bool)
	chIoStop := make(chan bool)
	chIoButtons := make(chan elevio.ButtonEvent)

	// Channels for virtual elevator
	chVirtualButtons := make(chan elevio.ButtonEvent)
	chVirtualFloor := make(chan int)

	chReAssign := make(chan map[string][][3]bool)
	chMsgFromFsm := make(chan elevator.Elevator)
	// ****** Go routines ******

	// Goroutine for local elevator
	go elevio.PollButtons(chIoButtons)
	go elevio.PollFloorSensor(chIoFloor)
	go elevio.PollObstructionSwitch(chIoObstical)
	go elevio.PollStopButton(chIoStop)

	// poll button press from other
	go bcast.Transmitter(udpRecover, chRecovElevToNet)
	go bcast.Receiver(udpRecover, chRecovElevFromNet)

	// Network
	go peers.Transmitter(udpPeer, id, chPeerTxEnable)
	go peers.Receiver(udpPeer, chPeerUpdate)

	go bcast.Transmitter(updData, chMsgToNetwork)
	go bcast.Receiver(updData, chMsgFromNetwork)

	go distributor.Distribute2(
		id,
		chReAssign,
		chMsgToNetwork,
		chVirtualButtons)

	eObj := boot.Elevator(id, port, chIoFloor)

	go reciver.Run(
		eObj,
		chIoButtons,
		chIoFloor,
		chIoObstical,
		chIoStop,
		chMsgFromNetwork,
		chReAssign,
		chVirtualFloor,
		chMsgFromFsm,
		chMsgToNetwork,
		chPeerUpdate)

	go FSM.FSM2(
		eObj,
		chVirtualButtons,
		chIoFloor,
		chIoObstical,
		chIoStop,
		chMsgFromFsm)

	// watchdog
	// not implement

	// distributor
	// not implement

	// can use empty select to block main thread
	select {}

}
