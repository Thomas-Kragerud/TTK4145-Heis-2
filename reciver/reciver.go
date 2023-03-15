package reciver

import "Project/elevio"

func run(
	chIoButtons <-chan elevio.ButtonEvent,
	chIoFloor <-chan int,
	chIoObstical <-chan bool,
	chIoStop <-chan bool) {

	select {
	case ioBtn := <-chIoButtons:

	case ioF := <-chIoFloor:

	case ioObst := <-chIoObstical:

	case ioS := <-chIoStop:

	}
}