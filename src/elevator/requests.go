package elevator

type DrinBehaviorPair struct {
	dirn     Dirn
	behavior ElevatorBehavior
}

func requests_above(e Elevator) bool {
	for f := e.floor + 1; f < N_FLOORS; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if /* e.requests[f][btn]*/ 1 == 1 {
				return true
			}
		}
	}
	return false
}

func requests_below(e Elevator) bool {
	for f := 0; f < e.floor; f++ {
		for btn := 0; btn < N_BUTTONS; btn++ {
			if e.requests[f][btn] == 1{
				return true
			}
		}
	}
	return false
}

func requests_here(e Elevator) bool {
	for btn := 0; btn < N_BUTTONS; btn++ {
		if e.requests[e.floor][btn] == 1 {
			return true
		}
	}
	return false
}

func requests_chooseDirection(e Elevator) DrinBehaviorPair {
	switch e.dirn {
	case D_Up:
		if requests_above(e)  {
			return DrinBehaviorPair{dirn: D_Up, behavior: EB_Moving}
		} else {
			if requests_here(e) {
				return DrinBehaviorPair{dirn: D_Down, behavior: EB_DoorOpen}
			} else {
				if requests_below(e) {
					return DrinBehaviorPair{dirn: D_Down, behavior: EB_Moving}
				} else {
					return DrinBehaviorPair{dirn: D_Stop, behavior: EB_Idle}
				}
			}
		}

	case D_Down:
		if requests_below(e) {
			return DrinBehaviorPair{dirn: D_Down, behavior: EB_Moving}
		} else {
			if requests_here(e) {
				return DrinBehaviorPair{dirn: D_Up, behavior: EB_DoorOpen}
			} else {
				if requests_above(e) {
					return DrinBehaviorPair{dirn: D_Up, behavior: EB_Moving}
				} else {
					return DrinBehaviorPair{dirn: D_Stop, behavior: EB_Idle}
				}
			}
		}

	case D_Stop:
		if requests_here(e) {
			return DrinBehaviorPair{dirn: D_Stop, behavior: EB_DoorOpen}
		} else {
			if requests_above(e) {
				return DrinBehaviorPair{dirn: D_Up, behavior: EB_Moving}
			} else {
				if requests_below(e) {
					return DrinBehaviorPair{dirn: D_Down, behavior: EB_Moving}
				} else {
					return DrinBehaviorPair{dirn: D_Stop, behavior: EB_Idle}
				}
			}
		}
	default:
		return DrinBehaviorPair{dirn: D_Stop, behavior: EB_Idle}
	}
}

func reguests_shouldStop(e Elevator) int {
	switch e.dirn {
	case D_Down:
		//return e.requests[e.floor][B_HallDown] || e.requests[e.floor][B_Cab] || !requests_below(e)
	}
	default:
		return 1
}
