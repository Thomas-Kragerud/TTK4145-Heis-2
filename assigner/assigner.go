package assigner

import (
	"Project/config"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
)

// Assign -
func Assign(data config.InputData) map[string][][]bool {
	// Convert the input data to JSON
	dataJson, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error encoding input data:", err)
		os.Exit(1)
	}
	// Define the command to execute the external executable with command line arguments
	cmd := exec.Command("costfunc/./hall_request_assigner",
		"--input", string(dataJson),
		"--travelDuration", "3000",
		"--doorOpenDuration", "4000",
		"--clearRequestType", "all",
		"--includeCab",
	)

	// Get the standard input and output streams of the child process
	stdin, err := cmd.StdinPipe()
	if err != nil {
		fmt.Println("Error getting standard input pipe:", err)
		os.Exit(1)
	}
	defer stdin.Close()

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("Error getting standard output pipe:", err)
		os.Exit(1)
	}

	// Start the child process
	if err := cmd.Start(); err != nil {
		fmt.Println("Error starting external executable:", err)
		os.Exit(1)
	}

	// Wait for the child process to complete
	if err := cmd.Wait(); err != nil {
		fmt.Println("External executable returned error:", err)
		os.Exit(1)
	}

	// Read the output data from the child process's standard output stream
	outputDataJson, err := io.ReadAll(stdout)
	if err != nil {
		fmt.Println("Error reading output data from standard output pipe:", err)
		os.Exit(1)
	}

	// Parse the output data as JSON and print it
	var outputData map[string][][]bool
	if err := json.Unmarshal(outputDataJson, &outputData); err != nil {
		fmt.Println("Error decoding output data:", err)
		os.Exit(1)
	}
	//var out map[string]config.SendHall
	//for id := range outputData {
	//	var h config.SendHall
	//	for f := range outputData[id] {
	//		for btn := range outputData[id][f] {
	//			h.HallRequests[f][btn] = outputData[id][f][btn]
	//		}
	//	}
	//	out[id] = h
	//}
	//return out
	return outputData
}
