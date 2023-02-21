package Elevator

type DrinBehaviorPair struct {
	dirn Dirn
	behavior ElevatorBehavior
}

func requests_above(e Elevator) int {
	for f := e.floor + 1; f < N_FLOORS; f++{
		for btn := 0; btn < N_BUTTONS; btn++ {
			if /* e.requests[f][btn]*/1 ==1 {
				return 1
			}
		}
	}
	return 0
}

func requests_below(e Elevator) int {
	for f := 0; f < e.floor; f++{
		for btn := 0; btn < N_BUTTONS; btn++ {
			if /* e.requests[f][btn]*/1 ==1 {
				return 1
			}
		}
	}
	return 0
}

func requests_here(e Elevator) int {
	for btn := 0; btn < N_BUTTONS; btn++{
		if e.requests[e.floor][btn] {
			return 1
		}
	}
	return 0
}

func requests_chooseDirection(e Elevator) DrinBehaviorPair{
	switch e.dirn {
	case D_UP:
		if requests_above(e) {
			return DrinBehaviorPair{dirn: D_Up, behavior: EB_Moving}
		} else {
			if requests_here(e) {
				return DrinBehaviorPair{dirn: D_Down, behavior: EB_DoorOpen}
			} else  {
				if requests_below(e) {
					return DrinBehaviorPair{dirn: D_Down, behavior: EB_Moving}
				} else {
					return DrinBehaviorPair{dirn: D_Stop, behavior: EB_Idle}
				}
			}
		}

	case D_DOWN:
		if requests_below(e) {
			return DrinBehaviorPair{dirn: D_Down, behavior: EB_Moving}
		} else {
			if requests_here(e) {
				return DrinBehaviorPair{dirn: D_Up, behavior: EB_DoorOpen}
			} else  {
				if requests_above(e) {
					return DrinBehaviorPair{dirn: D_Up, behavior: EB_Moving}
				} else {
					return DrinBehaviorPair{dirn: D_Stop, behavior: EB_Idle}
				}
			}
		}

	case D_STOP:
		if requests_here(e) {
			return DrinBehaviorPair{dirn: D_Stop, behavior: EB_DoorOpen}
		} else {
			if requests_above(e) {
				return DrinBehaviorPair{dirn: D_Up, behavior: EB_Moving}
			} else  {
				if requests_below(e) {
					return DrinBehaviorPair{dirn: D_Down, behavior: EB_Moving}
				} else {
					return DrinBehaviorPair{dirn: D_Stop, behavior: EB_Idle}
				}
			}
		}
	default:
		return DrinBehaviorPair{dirn: D_Stop,behavior: EB_Idle}
	}
}
