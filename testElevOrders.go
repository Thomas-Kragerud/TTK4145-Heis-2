package main

import (
	"Project/elevio"
	"Project/localElevator/elevator"
	"fmt"
	"Project/costfunc"
)

func main() {
	e1 := new(elevator.Elevator)
	e1.Id = "1"
	e1.Floor = 0
	e1.Dir = elevio.MotorDirection(2)
	e1.Orders = make([][]bool, 4)
	e1.Orders = [][]bool{{true, false, false}, {false, false, false}, {false, false, false}, {false, false, false}}
	e1.State = elevator.ElevatorState(0)
	e1.Obs = false

	e2 := new(elevator.Elevator)
	e2.Id = "2"
	e2.Floor = 3
	e2.Dir = elevio.MotorDirection(2)
	e1.Orders = [][]bool{{false, false, false}, {false, false, false}, {false, false, false}, {false, true, false}}
	e2.State = elevator.ElevatorState(0)
	e2.Obs = false

	list_of_active_elevators := []*elevator.Elevator{e1, e2}
	fmt.Println(list_of_active_elevators)

	input := costfunc.create_input(list_of_active_elevators)

	costfunc.runCostfunc(input)

}
