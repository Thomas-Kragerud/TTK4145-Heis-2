package reciver

import (
	"Project/assigner"
	"Project/config"
	"Project/elevio"
	"fmt"
)

func Run(
	pid string,
	chIoButtons <-chan elevio.ButtonEvent,
	chIoFloor <-chan int,
	chIoObstical <-chan bool,
	chIoStop <-chan bool) {

	// Init this elevator i main
	var thisElev config.SendElev
	allStates := make(map[string]config.SendElev)
	var hR config.SendHall

	var data config.InputData
	output := make(map[string][][]bool)

	thisElev.Init()
	hR.Init()
	allStates[pid] = thisElev

	data.HallRequests.HallRequests = hR.HallRequests
	data.States = allStates

	// Init disse

	for {
		select {
		case ioBtn := <-chIoButtons:
			if ioBtn.Button == elevio.BT_Cab {
				allStates[pid].CabRequests[ioBtn.Floor] = true
				data.HallRequests = hR
				// Formater til input data objekt og send til assigner

				fmt.Printf("Recived cat call at %d\n", ioBtn.Floor)
			} else {
				hR.Update(ioBtn.Floor, ioBtn.Button)
				fmt.Printf("Recived hallbtn at %d", ioBtn.Floor)
				// Noe om at den ikke er assigned
			}
			data.HallRequests.HallRequests = hR.HallRequests
			data.States = allStates
			output = assigner.Assign(data)
			fmt.Printf("Output: %v\n", output)

			break // Trenger jeg ???

		case ioF := <-chIoFloor:
			thisElev.Floor = ioF
			allStates[pid] = thisElev
			fmt.Printf("Recived new floor %d", ioF)

			//case ioObst := <-chIoObstical:
			//
			//
			//case ioS := <-chIoStop:
			//
			//case card := <-chBrodcastet:

		}
	}
}
