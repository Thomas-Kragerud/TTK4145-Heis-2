package distributor

import (
	"Project/elevio"
	"Project/localElevator/elevator"
)

func Distribute2(
	pid string,
	chReAssign <-chan map[string][][3]bool,
	chMsgToNetwork chan<- elevator.Elevator,
	chVirtualButtons chan<- elevio.ButtonEvent,
	chRemoveOrders chan<- elevio.ButtonEvent) {

	for {
		select {
		case reAssign := <-chReAssign:
			//fmt.Printf("Reassigning orders: %v\n", reAssign)
			go func() {
				for id, orders := range reAssign {
					if id == pid {
						for f := range orders {
							for b := range orders[f] {
								if orders[f][b] {
									chVirtualButtons <- elevio.ButtonEvent{
										Floor:  f,
										Button: elevio.ButtonType(b)}
								} else {
									chRemoveOrders <- elevio.ButtonEvent{
										Floor:  f,
										Button: elevio.ButtonType(b)}
								}
							}
						}
					}
				}
			}()

		default:
			continue
		}

	}
}
