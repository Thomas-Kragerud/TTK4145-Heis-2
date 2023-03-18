package main

import (
	"Project/assigner"
	"Project/config"
)
import "fmt"

// Struct members must be public in order to be accessible by json.Marshal/.Unmarshal
// This means they must start with a capital letter, so we need to use field renaming struct tags to make them camelCase

type HRAElevState struct {
	Behavior    string `json:"behaviour"`
	Floor       int    `json:"floor"`
	Direction   string `json:"direction"`
	CabRequests []bool `json:"cabRequests"`
}

type HRAInput struct {
	HallRequests [][2]bool               `json:"hallRequests"`
	States       map[string]HRAElevState `json:"states"`
}

func main() {

	input := config.HRAInput{
		HallRequests: [][2]bool{{false, false}, {true, false}, {false, false}, {false, true}},
		States: map[string]config.HRAElevState{
			"one": config.HRAElevState{
				Behavior:    "moving",
				Floor:       2,
				Direction:   "up",
				CabRequests: []bool{false, false, false, true},
			},
			"two": config.HRAElevState{
				Behavior:    "idle",
				Floor:       0,
				Direction:   "stop",
				CabRequests: []bool{false, false, false, false},
			},
		},
	}

	//jsonBytes, err := json.Marshal(input)
	//if err != nil {
	//	fmt.Println("json.Marshal error: ", err)
	//	return
	//}
	//
	//ret, err := exec.Command("docker", "run", "--rm", "-i", "dock_hra", "/app/hall_request_assigner", "--input", string(jsonBytes)).CombinedOutput()

	output := assigner.Assign(input)
	//output := new(map[string][][2]bool)
	//err = json.Unmarshal(ret, &output)
	//if err != nil {
	//	fmt.Println("json.Unmarshal error: ", err)
	//	return
	//}
	//t := reflect.TypeOf(err)
	//fmt.Println("Type of myVar:", t)

	fmt.Printf("output: \n")
	for k, v := range output {
		fmt.Printf("%6v :  %+v\n", k, v)
	}

}
