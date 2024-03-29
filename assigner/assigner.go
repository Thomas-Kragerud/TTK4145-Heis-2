package assigner

import (
	"Project/config"
	"encoding/json"
	"fmt"
	"os/exec"
	"runtime"
)


// This function takes a input on the HRAI form, and returns a map of the new orders of all the elevators present in the HRAI. That is, this elevator 
// reassignes all the active orders using the executable provides. The executable is located in the costfunc folder
func Assign(inData config.HRAInput) map[string][][3]bool {
	jsonBytes, err := json.Marshal(inData)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
	}

	var ret []uint8 

	if runtime.GOOS == "darwin" {
		ret, err = exec.Command(
			"docker",
			"run",
			"--rm",
			"-i",
			"dock_hra",
			"/app/hall_request_assigner",
			"--input",
			string(jsonBytes)).CombinedOutput()
	} else if runtime.GOOS == "windows" {
		ret, err = exec.Command(
			"costfunc/hall_request_assigner.exe",
			"-i",
			string(jsonBytes)).CombinedOutput()
	} else if runtime.GOOS == "linux" {
		ret, err = exec.Command(
			"costfunc/hall_request_assigner",
			"-i",
			string(jsonBytes)).CombinedOutput()
	}

	output := new(map[string][][3]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
		fmt.Println("Command output:", string(ret))
		fmt.Printf("input Data %v\n", inData)
	}
	return *output

}
