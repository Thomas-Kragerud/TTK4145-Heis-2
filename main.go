package main

import (
	"Project/elevio"
	"Project/localElevator/FSM"
	"Project/localElevator/boot"
	"Project/localElevator/elevator"
	"Project/network/bcast"
	"Project/network/peers"
	"Project/networkHandler"
	"Project/reciver"
	"flag"
)

func main() {
	// parameter setting
	var udpPeer = 6000
	var updData = 6200
	var udpRecover = 6400
	var udpAddBtn = 6450
	var udpRmBtn = 6500
	var id string
	var port string

	flag.StringVar(&port, "port", "", "Port of this elevator")
	flag.StringVar(&id, "id", "", "id of this elevator")
	flag.Parse()

	// ****** Set up channels ******

	// Channels for distribution
	//chMsgToNetwork := make(chan elevator.Elevator)
	//chMsgFromNetwork := make(chan elevator.Elevator)
	chMsgToNetwork := make(chan networkHandler.NetworkPackage)
	chMsgFromNetwork := make(chan networkHandler.NetworkPackage)

	chRecovElevToNet := make(chan elevator.Elevator)
	chRecovElevFromNet := make(chan elevator.Elevator)

	chPeerUpdate := make(chan peers.PeerUpdate)
	chPeerTxEnable := make(chan bool)

	// Channels for local elevator
	chIoFloor := make(chan int)
	chIoObstical := make(chan bool)
	chIoStop := make(chan bool)
	chIoButtons := make(chan elevio.ButtonEvent, 100)

	// Channels for virtual elevator
	chVirtualButtons := make(chan elevio.ButtonEvent, 100)
	chRemoveOrders := make(chan elevio.ButtonEvent, 100)
	chVirtualFloor := make(chan int)

	chReAssign := make(chan map[string][][3]bool, 100)
	chMsgFromFsm := make(chan elevator.Elevator, 100)
	//chMsgToFsm := make(chan elevator.Elevator)
	// ****** Go routines ******
	chAddBtnNet := make(chan elevio.ButtonEvent)
	chReciveBtnNet := make(chan elevio.ButtonEvent, 100)
	chRmBtnNet := make(chan elevio.ButtonEvent, 100)
	chRmReciveBtnNet := make(chan elevio.ButtonEvent, 100)
	chClareHallFsm := make(chan elevio.ButtonEvent, 100)

	// Goroutine for local elevator
	go elevio.PollButtons(chIoButtons)
	go elevio.PollFloorSensor(chIoFloor)
	go elevio.PollObstructionSwitch(chIoObstical)
	go elevio.PollStopButton(chIoStop)

	// poll button press from other
	go bcast.Transmitter(udpRecover, chRecovElevToNet)
	go bcast.Receiver(udpRecover, chRecovElevFromNet)

	go bcast.Transmitter(udpAddBtn, chAddBtnNet)
	go bcast.Receiver(udpAddBtn, chReciveBtnNet)
	go bcast.Transmitter(udpRmBtn, chRmBtnNet)
	go bcast.Receiver(udpRmBtn, chRmReciveBtnNet)

	// Network
	go peers.Transmitter(udpPeer, id, chPeerTxEnable)
	go peers.Receiver(udpPeer, chPeerUpdate)

	go bcast.Transmitter(updData, chMsgToNetwork)
	go bcast.Receiver(updData, chMsgFromNetwork)

	//go distributor.Distribute2(
	//	id,
	//	chReAssign,
	//	chMsgToNetwork,
	//	chMsgToFsm,
	//	chVirtualButtons,
	//	chRemoveOrders)

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
		chPeerUpdate,
		chAddBtnNet,
		chRmBtnNet,
		chClareHallFsm,
		chReciveBtnNet,
		chRmReciveBtnNet)

	go FSM.FSM2(
		eObj,
		chVirtualButtons,
		chIoFloor,
		chIoObstical,
		chIoStop,
		chMsgFromFsm,
		chRemoveOrders,
		chReAssign,
		chClareHallFsm)

	// watchdog
	// not implement

	// distributor
	// not implement

	// can use empty select to block main thread
	select {}

}
