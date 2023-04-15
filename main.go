package main

import (
	"Project/config"
	"Project/elevio"
	"Project/localElevator/boot"
	"Project/localElevator/elevator"
	"Project/localElevator/fsm"
	"Project/messageHandler"
	"Project/network/versionController"
	"Project/network/bcast"
	"Project/network/peers"
	"flag"
)

/* func main() {
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
	chMsgToNetwork := make(chan versionController.SROnNet, 100) // Buffer sånn at får med alle button press selv om kanskje dæver(?)
	chMsgFromNetwork := make(chan versionController.SROnNet, 100)

	chMsgToSend := make(chan messageHandler.NetworkPackage, 100)
	chMsgFromReciver := make(chan messageHandler.NetworkPackage, 100)

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

	go versionController.Send(
		chMsgToNetwork,
		chMsgToSend,
	)

	go versionController.Recive(
		chMsgFromNetwork,
		chMsgFromReciver, )

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

} */

func main() {
	var id string
	var port string
	udpPeer := 6001
	udpData := 6002
	var numFloors int

	flag.StringVar(&port, "port", "", "Port of this elevator")
	flag.StringVar(&id, "id", "", "id of this elevator")
	flag.IntVar(&numFloors, "floors", 4, "number of elevator floors")
	flag.Parse()
	config.NumFloors = numFloors // Update config

	// Channels for networkMessaging
	chIoButtons := make(chan elevio.ButtonEvent, 1)
	msgToNetwork := make(chan versionController.SROnNet, 100) // Buffer sånn at får med alle button press selv om kanskje dæver(?)
	msgFromNetwork := make(chan versionController.SROnNet, 100)



	//Message channels to VersionControl Send and from VersionControl Reciever
	msgToVCSend := make(chan messageHandler.NetworkPackage, 100)
	msgFromVCReciver := make(chan messageHandler.NetworkPackage, 100)


	//Channels for network
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

	go bcast.Transmitter(udpData, msgToNetwork)
	go bcast.Receiver(udpData, msgFromNetwork)

	
	//Version control
	go versionController.VCSend(
		msgToNetwork,
		msgToVCSend,
	)

	go versionController.VCRecive(
		msgFromNetwork,
		msgFromVCReciver, )

	
	//FSM
	go fsm.Fsm(
		&eObj,
		chIoFloor,
		chIoObstical,
		chIoStop,
		chNewState,
		chRmButton,
		chAddButton,
	)

	//MessageHandler
	go messageHandler.Handle(
		&eObjCopy,
		chIoButtons,
		msgFromVCReciver,
		msgToVCSend,
		chNewState,
		chAddButton,
		chRmButton,
		chPeerUpdate,
	)
	select {}

}
