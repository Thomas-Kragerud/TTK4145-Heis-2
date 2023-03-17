package main

import (
	"Project/distributor"
	"Project/elevio"
	"Project/localElevator/FSM"
	"Project/localElevator/elevator"
	"Project/network/bcast"
	"Project/network/peers"
	//"flag"
)

//NILS VAR HER!

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

	// Channels for communication between distributor and watchdog
	// not implement

	// Channels for communication between distributor and a single elevator
	// not implement

	// Channels for local elevator
	chAtFloor := make(chan int)
	chObst := make(chan bool)
	chStop := make(chan bool)
	chButtons := make(chan elevio.ButtonEvent)

	// ****** Go routines ******

	// Goroutine for local elevator
	go elevio.PollButtons(chButtons)
	go elevio.PollFloorSensor(chAtFloor)
	go elevio.PollObstructionSwitch(chObst)
	go elevio.PollStopButton(chStop)

	// poll button press from other
	go bcast.Transmitter(udpRecover, chRecovElevToNet)
	go bcast.Receiver(udpRecover, chRecovElevFromNet)

	// Network
	go peers.Transmitter(udpPeer, id, chPeerTxEnable)
	go peers.Receiver(udpPeer, chPeerUpdate)

	go bcast.Transmitter(updData, chMsgToNetwork)
	go bcast.Receiver(updData, chMsgFromNetwork)

	go distributor.Distribute(id,
		chButtons,
		chMsgFromNetwork,
		chPeerUpdate,
		chRecovElevToNet,
		chRecovElevFromNet)

	// Go fms
	go FSM.FSM(
		port,
		id,
		chMsgToNetwork,
		chMsgFromNetwork,
		chButtons,
		chAtFloor,
		chObst,
		chStop)

	// watchdog
	// not implement

	// distributor
	// not implement

	// can use empty select to block main thread
	select {}
}
