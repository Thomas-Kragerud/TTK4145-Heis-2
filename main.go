package main

import (
	"Project/network/bcast"
	"Project/network/peers"
	"Project/singleElevator/elevator"
	"Project/singleElevator/elevio"
	"Project/singleElevator/singleFSM"
	"fmt"
	"time"
)

func main() {
	// parameter setting
	var port = "8080"
	var udpPeer = 6000
	var updData = 6200
	var udpRecover = 6400

	// ****** Set up channels ******

	// Channels for distribution
	chButtons := make(chan elevio.ButtonEvent)
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

	// ****** Go routines ******

	// Goroutine for local elevator
	go elevio.PollButtons(chButtons)
	go elevio.PollFloorSensor(chAtFloor)
	go elevio.PollObstructionSwitch(chObst)
	go elevio.PollStopButton(chStop)

	// poll button press from other
	go bcast.Transmitter(udpRecover, chRecovElevToNet)
	go bcast.Receiver(udpRecover, chRecovElevFromNet)
	go func() {
		for {
			select {
			case recovElevat := <-chRecovElevFromNet:
				fmt.Printf("Forsøk og recover\n")
				e := recovElevat
				for f := range e.Orders {
					for btn := range e.Orders[f] {
						fmt.Printf("Looped \n")
						if e.Orders[f][btn] {
							fmt.Printf("Sender gamle states \n")
							chButtons <- elevio.ButtonEvent{
								Floor:  f,
								Button: elevio.ButtonType(int(btn))}
							time.Sleep(50 * time.Millisecond)
						}

					}
				}
			}
		}
	}()

	// Network
	var id = "69"
	go peers.Transmitter(udpPeer, id, chPeerTxEnable)
	go peers.Receiver(udpPeer, chPeerUpdate)

	go bcast.Transmitter(updData, chMsgToNetwork)
	go bcast.Receiver(updData, chMsgFromNetwork)

	// Go fms
	go singleFSM.FSM(
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
