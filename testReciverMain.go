package main

import (
	"Project/elevio"
	"flag"
)

func main() {
	// parameter setting
	//var udpPeer = 6000
	//var updData = 6200
	//var udpRecover = 6400
	var id string
	var port string

	flag.StringVar(&port, "port", "", "Port of this elevator")
	flag.StringVar(&id, "id", "", "id of this elevator")
	flag.Parse()

	chAtFloor := make(chan int)
	chObst := make(chan bool)
	chStop := make(chan bool)
	chButtons := make(chan elevio.ButtonEvent)

	go elevio.PollButtons(chButtons)
	go elevio.PollFloorSensor(chAtFloor)
	go elevio.PollObstructionSwitch(chObst)
	go elevio.PollStopButton(chStop)

	elevio.Init("localhost:"+port, 4)

	select {}

}
