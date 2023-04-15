// Note to self. In elevator_io.go we use 3 as button
// Here we use config.NumButtons. Update to enforce consistency

package elevator

import (
	"Project/config"
	"Project/elevio"
	"fmt"
)

type Elevatorstate int

const (
	Idle     Elevatorstate = 0
	DoorOpen Elevatorstate = 1
	Moving   Elevatorstate = 2
)

var _stateToString = map[Elevatorstate]string{
	Idle:     "idle",
	DoorOpen: "doorOpen",
	Moving:   "moving",
}

type Elevator struct {
	Floor        int
	Dir          elevio.MotorDirection
	Orders       [][]bool
	State        Elevatorstate
	Id           string
	Obs          bool // Obstruction
	ReAssignStop bool
	//OrderMutex sync.Mutex // Add a mutex to the Elevator struct
}

// Init initialize the elevator
func (e *Elevator) Init(Id string) {
	// Create the order matrix
	e.Orders = make([][]bool, config.NumFloors)
	for floor := range e.Orders {
		e.Orders[floor] = make([]bool, config.NumButtons)
	}
	// Set the rest of the parameters
	e.Id = Id
	e.Floor = 0
	e.State = Idle
	e.Dir = elevio.MD_Stop
	e.Obs = false
	e.ReAssignStop = false
}

func (e *Elevator) Clone() Elevator {
	// Create a deep copy of the Orders field
	clonedOrders := make([][]bool, len(e.Orders))
	for i := range e.Orders {
		clonedOrders[i] = make([]bool, len(e.Orders[i]))
		copy(clonedOrders[i], e.Orders[i])
	}

	return Elevator{
		State:  e.State,
		Floor:  e.Floor,
		Dir:    e.Dir,
		Orders: clonedOrders,
		Id:     e.Id,
		Obs:    e.Obs,
		//OrderMutex: sync.Mutex{}, // Create a new mutex for the cloned object
	}
}

// String toString method for elevator object
func (e *Elevator) String() string {
	str := "***Elevator***\n"
	str += fmt.Sprintf("Floor: %d\n", e.Floor)
	str += fmt.Sprintf("Direction: %s\n", elevio.ToStringMotorDirection(e.Dir))
	str += fmt.Sprintf("State: %s\n", _stateToString[e.State])
	str += "Orders:\n"
	str += "  f hUp   hDown   cab\n"
	for floor := range e.Orders {
		str += fmt.Sprintf("| %d", floor)
		for _, btn := range e.Orders[floor] {
			str += fmt.Sprintf(" %v ", btn)
		}
		str += "|\n"
	}
	return str
}

// UpdateLights - updates all lights except the hall lights
//
//	which are set in the messageHandler.go event loop
func (e *Elevator) UpdateLights() {
	elevio.SetFloorIndicator(e.Floor)
	for floor := range e.Orders {
		elevio.SetButtonLamp(elevio.ButtonType(elevio.BT_Cab), floor, e.Orders[floor][elevio.BT_Cab])
	}
}

// AddOrder - adds a ButtonEvent to the Order matrix
func (e *Elevator) AddOrder(event elevio.ButtonEvent) {
	e.Orders[event.Floor][event.Button] = true
}

// ClearOrderAtFloor clears all orders at the specified floor in the elevator's order matrix.
func (e *Elevator) ClearOrderAtFloor(floor int) {
	e.Orders[floor][elevio.BT_Cab] = false
	if e.Dir == elevio.MD_Up {
		e.Orders[floor][elevio.BT_HallUp] = false

		if e.Orders[floor][elevio.BT_HallDown] {
			e.Orders[floor][elevio.BT_HallDown] = false 
		}
	} else if e.Dir == elevio.MD_Down {
		e.Orders[floor][elevio.BT_HallDown] = false

		if e.Orders[floor][elevio.BT_HallUp] {
			e.Orders[floor][elevio.BT_HallUp] = false
		}
	}
}

func (e *Elevator) ClearOrderFromBtn(button elevio.ButtonEvent) {
	//e.OrderMutex.Lock()
	//defer e.OrderMutex.Unlock()
	e.Orders[button.Floor][button.Button] = false
}

// ClearAllOrders removes all orders from the elevator's order matrix
// and turns off all associated button lamps.
func (e *Elevator) ClearAllOrders() {
	for f := 0; f < config.NumFloors; f++ {
		e.ClearOrderAtFloor(f)
		for b := elevio.ButtonType(0); b < 3; b++ {
			elevio.SetButtonLamp(b, f, false)
			//fmt.Println("Clearing all lights")
		}
	}
}

func (e *Elevator) OrderIsEmpty() bool {
	for f := range e.Orders {
		for btn := range e.Orders[f] {
			if e.Orders[f][btn] {
				return false
			}
		}
	}
	return true
}

// ** Trenger jeg disse funksjone fra et kodekvali perspektiv?? **

// Delete when deleting old fsm
func (e *Elevator) SetDirectionStop() {
	e.Dir = elevio.MD_Stop
}

func (e *Elevator) SetFloor(floor int) {
	e.Floor = floor
}

func (e *Elevator) SetStateDoorOpen() {
	e.State = DoorOpen
}

func (e *Elevator) SetStateIdle() {
	e.State = Idle
}

func (e *Elevator) SetStateMoving() {
	e.State = Moving
}

func (e *Elevator) ToHRA() config.HRAElevState {
	//e.OrderMutex.Lock()         // Lock the mutex before modifying the Orders field
	//defer e.OrderMutex.Unlock() // Defer unlocking the mutex, so it's released even if the function returns early
	var cabReq []bool
	for _, btn := range e.Orders {
		cabReq = append(cabReq, btn[2])
	}
	return config.HRAElevState{
		Behavior:    _stateToString[e.State],
		Floor:       e.Floor,
		Direction:   elevio.ToStringMotorDirection(e.Dir),
		CabRequests: cabReq,
	}
}

func (e *Elevator) ToHallReq() [][2]bool {
	var hallReq [][2]bool
	for _, floor := range e.Orders {
		hallReq = append(hallReq, [2]bool{floor[0], floor[1]})
	}
	fmt.Printf("HallReq: %v\n", hallReq)
	return hallReq
}
