package Elevator

/*

Must fix import og fil greier
Har ikke dirnGo first searches for package directory inside GOROOT/src directory and if it doesn’t find the package, then it looks for GOPATH/src. Since, fmt package is part of Go’s standard library which is located in GOROOT/src, it is imported from there. But since Go cannot find greet package inside GOROOT, it will lookup inside GOPATH/src and we have it there. og reguests greiene

*/

type ElevatorBehavior string

const (
	EB_Idle ElevatorBehavior = "EB_Idle"
	EB_DoorOpen ElevatorBehavior = "EB_DoorOpen"
	EB_Moving ElevatorBehavior = "EB_Moving"
)

type ClearRequestVariant string

const (
	CV_ALL ClearRequestVariant = "all"
	CV_InDir ClearRequestVariant = "inDir"
)

type Config struct {
	clearRequestVariant ClearRequestVariant
	doorOpenDuration uint32 
}

type Elevator struct {
	floor int
	//dirn Dirn
	//requests int-list
	behavior ElevatorBehavior
	config Config
}

//Might ad later

/*
func elevator_print(es Elevator) {

}

func eb_toStrinng(ElevatorBehavior eb) {

}
*/  

func elevator_uninitialized() Elevator {
	config := Config{clearRequestVariant: CV_ALL, doorOpenDuration: 3.0}
	elevator := Elevator{floor: -1,
						/*dirn: D_stop, 
						behavior: EB_Idle,*/
						config: config}
	return elevator 
}

