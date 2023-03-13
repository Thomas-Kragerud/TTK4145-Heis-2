// Note to self. In elevator_io.go we use 3 as button
// Here we use config.NumButtons. Update to enforce consistency

package elevator

import (
	"Project/config"
	"Project/singleElevator/elevio"
	"fmt"
)

type elevatorState int

const (
	Idle     elevatorState = 0
	DoorOpen elevatorState = 1
	Moving   elevatorState = 2
)

var _stateToString = map[elevatorState]string{
	Idle:     "Idle",
	DoorOpen: "DoorOpen",
	Moving:   "Moving",
}

type ElevatorInterface interface {
	//Init()
	String() string
	//UpdateLights()
	//AddOrder(event elevio.ButtonEvent)
	//SetDirectionDown()
	//SetDirectionUp()
	//SetDirectionStop()
	//SetFloor(floor int)
}

type Elevator struct {
	Floor  int
	Dir    elevio.MotorDirection
	Orders [][]bool
	State  elevatorState
	Id     string
	Obs    bool // Obstruction
}

// Init initialize the elevator
func (e *Elevator) Init(Id string, chOldElevator <-chan Elevator) {
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

// UpdateLights oppdaterer alle knappene til heisen
func (e *Elevator) UpdateLights() {
	elevio.SetFloorIndicator(e.Floor)
	for floor := range e.Orders {
		for b := elevio.ButtonType(0); b < 3; b++ {
			elevio.SetButtonLamp(b, floor, e.Orders[floor][b])
		}
	}
}

// *** Functions for local elevator order list ***

func (e *Elevator) AddOrder(event elevio.ButtonEvent) {
	e.Orders[event.Floor][event.Button] = true
}

func (e *Elevator) ClearOrderAtFloor(floor int) {
	for btn, _ := range e.Orders[floor] {
		e.Orders[floor][btn] = false
	}
}

func (e *Elevator) ClearAllOrders() {
	for f := 0; f < config.NumFloors; f++ {
		e.ClearOrderAtFloor(f)
	}
}

func (e *Elevator) ClearAtCurrentFloor() {
	for btn, _ := range e.Orders[e.Floor] {
		e.Orders[e.Floor][btn] = false
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

// SetDirectionDown sets the motor direction to down
func (e *Elevator) SetDirectionDown() {
	e.Dir = elevio.MD_Down
}

func (e *Elevator) SetDirectionUp() {
	e.Dir = elevio.MD_Up
}

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
