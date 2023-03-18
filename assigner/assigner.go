package assigner

import (
	"Project/config"
	"encoding/json"
	"fmt"
	"os/exec"
)

// Assign -
func Assign(inData config.HRAInput) map[string][][2]bool {
	// Mekk test asigner inn her og få den til å funke med
	jsonBytes, err := json.Marshal(inData)
	if err != nil {
		fmt.Println("json.Marshal error: ", err)
	}
	ret, err := exec.Command(
		"docker",
		"run",
		"--rm",
		"-i",
		"dock_hra",
		"/app/hall_request_assigner",
		"--input",
		string(jsonBytes)).CombinedOutput()

	output := new(map[string][][2]bool)
	err = json.Unmarshal(ret, &output)
	if err != nil {
		fmt.Println("json.Unmarshal error: ", err)
	}

	return *output

	// Convert the input data to JSON

}
