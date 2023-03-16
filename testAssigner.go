package main

import (
	"os/exec"
)
import "fmt"
import "encoding/json"

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

	input := HRAInput{
		HallRequests: [][2]bool{{false, false}, {true, false}, {false, false}, {false, true}},
		States: map[string]HRAElevState{
			"one": HRAElevState{
				Behavior:    "moving",
				Floor:       2,
				Direction:   "up",
				CabRequests: []bool{false, false, false, true},
			},
			"two": HRAElevState{
				Behavior:    "idle",
				Floor:       0,
				Direction:   "stop",
				CabRequests: []bool{false, false, false, false},
			},
		},
	}

	jsonBytes, err := json.Marshal(input)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
		return
	}

	//ret, err := exec.Command("docker", "run", "--rm", "dock_hra", "-i", string(jsonBytes)).CombinedOutput()
	//if err != nil {
	//	fmt.Println("exec.Command error: ", err, "\nOutput:", string(ret))
	//
	//	//fmt.Println(string(ret))
	//	return
	//}
	//cmd := exec.Command("docker", "run", "--rm", "-i", "dock_hra", "--input", string(jsonBytes))
	ret, err := exec.Command("docker", "run", "--rm", "-i", "dock_hra", "/app/hall_request_assigner", "--input", string(jsonBytes)).CombinedOutput()

	//var output bytes.Buffer
	//cmd.Stdout = &output
	//cmd.Stderr = &output
	//
	//err = cmd.Run()
	//if err != nil {
	//	fmt.Println("exec.Command error: ", err, "\nOutput:", output.String())
	//	return
	//}

	//outputs := new(map[string][][2]bool)
	//err = json.Unmarshal(output.Bytes(), &output)
	//if err != nil {
	//	fmt.Println("json.Unmarshal error: ", err)
	//	return
	//}
	//fmt.Printf("%v", cmd)

	output := new(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		return
	}

	fmt.Printf("output: \n")
	for k, v := range *output {
		fmt.Printf("%6v :  %+v\n", k, v)
	}

}
