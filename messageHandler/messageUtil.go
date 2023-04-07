package messageHandler

import (
	"Project/assigner"
	"Project/config"
	"Project/elevio"
)

// reAssign
// Reassigns all hall orders
// and returns a slice of button events that later will update the FSM
func reAssign(
	pid string,
	elevatorMap map[string]ElevatorUpdate,
	hall [][2]bool) []assignValue {

	input := config.HRAInput{
		States:       make(map[string]config.HRAElevState),
		HallRequests: make([][2]bool, config.NumFloors)}
	for id, val := range elevatorMap {
		if val.Alive {
			hraElev := val.Elevator.ToHRA()
			input.States[id] = hraElev
		}
	}
	input.HallRequests = hall
	result := assigner.Assign(input)
	hallBefore := elevatorMap[pid].Elevator.Orders
	hallAfter := result[pid]
	fromReAssigner := make([]assignValue, 0)

	for f := 0; f < config.NumFloors; f++ {
		for b := elevio.ButtonType(0); b < 2; b++ {
			if hallAfter[f][b] && !hallBefore[f][b] {
				fromReAssigner = append(
					fromReAssigner,
					assignValue{Add, elevio.ButtonEvent{f, b}})
			}
			if !hallAfter[f][b] && hallBefore[f][b] {
				fromReAssigner = append(
					fromReAssigner,
					assignValue{Remove, elevio.ButtonEvent{f, b}})
			}
		}
	}
	return fromReAssigner
}

func addHallBTN(hall [][2]bool, btn elevio.ButtonEvent) [][2]bool {
	if btn.Button == elevio.BT_HallUp {
		hall[btn.Floor][0] = true
	} else if btn.Button == elevio.BT_HallDown {
		hall[btn.Floor][1] = true
	}
	return hall
}

func clareHallBTN(hall [][2]bool, btn elevio.ButtonEvent) [][2]bool {
	if btn.Button == elevio.BT_HallUp {
		hall[btn.Floor][0] = false
	} else if btn.Button == elevio.BT_HallDown {
		hall[btn.Floor][1] = false
	}
	return hall
}
func updateHallLights(hall [][2]bool) {
	for f := range hall {
		if hall[f][0] {
			elevio.SetButtonLamp(elevio.BT_HallUp, f, true)
		} else {
			elevio.SetButtonLamp(elevio.BT_HallUp, f, false)
		}
		if hall[f][1] {
			elevio.SetButtonLamp(elevio.BT_HallDown, f, true)
		} else {
			elevio.SetButtonLamp(elevio.BT_HallDown, f, false)
		}
	}
}
