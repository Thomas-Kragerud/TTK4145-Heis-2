// Package for interacting with the io of the elevator
// all global variables defined inn this package is used only inside this package
//
// The types MotorDirection, ButtonType and ButtonEvent, are defined here, and are used throughout the program

package elevio

import "time"
import "sync"
import "net"
import "fmt"

// Structs and consts
const _pollRate = 20 * time.Millisecond

var _initialized bool = false
var _numFloors int = 4
var _mtx sync.Mutex
var _conn net.Conn // Connection to elevator

// MotorDirection The direction the elevator is traveling
type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

// _motorDirectionToString Modified to work with assigner format
var _motorDirectionToString = map[MotorDirection]string{
	MD_Up:   "up",
	MD_Down: "down",
	MD_Stop: "stop",
}

// ButtonType The different button presses
type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

// ButtonEvent type of button and at which floor
type ButtonEvent struct {
	Floor  int
	Button ButtonType
}

// Init initialise the global variables used throughout the program
func Init(addr string, numFloors int) {
	if _initialized {
		fmt.Println("Driver already initialized!")
		return
	}
	_numFloors = numFloors
	_mtx = sync.Mutex{}
	var err error
	_conn, err = net.Dial("tcp", addr)
	if err != nil {
		panic(err.Error())
	}
	_initialized = true
}

// PollButtons
// takes inn a send-only chanel (receiver) of type ButtonEvent
// for _pollRate loop over all the floors and check if any of the buttons has changed
// if any button has changed to true, send that over the channel receiver
// The function runs throughout the length of the program
func PollButtons(receiver chan<- ButtonEvent) {
	prev := make([][3]bool, _numFloors)
	for {
		time.Sleep(_pollRate)
		for f := 0; f < _numFloors; f++ {
			for b := ButtonType(0); b < 3; b++ {
				v := GetButton(b, f)               // True if pressed
				if v != prev[f][b] && v != false { // if not same as previous and not false, send button event to receiver channel
					receiver <- ButtonEvent{f, ButtonType(b)}
				}
				prev[f][b] = v
			}
		}
	}
}

// PollFloorSensor
// takes in a send-only channel (receiver) of type int
// The function sends the current floor over the channel when the floor changes
// When the floor does not change the function sends nothing
func PollFloorSensor(receiver chan<- int) {
	prev := -1
	for {
		time.Sleep(_pollRate)
		v := GetFloor()
		if v != prev && v != -1 {
			receiver <- v
		}
		prev = v
	}
}

// PollStopButton
// The stop-button is not one of the button types. Thus, it is not updated as buttonEvents
// The function takes in a send-only channel of type bool.
// If the value of the button changes, function sends true for active stopbutton, false for not
func PollStopButton(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetStop()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

// PollObstructionSwitch
// The Obstruction switch is like the stop-button not one of the standard button types.
// The function takes in a send-only channel of type bool.
// If the state of the obstruction switch changes, this change is transmitted on the receiver channel
func PollObstructionSwitch(receiver chan<- bool) {
	prev := false
	for {
		time.Sleep(_pollRate)
		v := GetObstruction()
		if v != prev {
			receiver <- v
		}
		prev = v
	}
}

//Functions: abstractions over the hardware/simulator

// SetMotorDirection Sets the direction of the elevator
func SetMotorDirection(dir MotorDirection) {
	write([4]byte{1, byte(dir), 0, 0})
}

// SetButtonLamp light up the correct button upon button event
func SetButtonLamp(button ButtonType, floor int, value bool) {
	write([4]byte{2, byte(button), byte(floor), toByte(value)})
}

// SetFloorIndicator light up the current floor of the elevator
func SetFloorIndicator(floor int) {
	write([4]byte{3, byte(floor), 0, 0})
}

// SetDoorOpenLamp sets the door open lamp :))
func SetDoorOpenLamp(value bool) {
	write([4]byte{4, toByte(value), 0, 0})
}

// SetStopLamp sets the stop lamp :))
func SetStopLamp(value bool) {
	write([4]byte{5, toByte(value), 0, 0})
}

// Not sure if functions below are needed outside this file?

// GetButton returns true if button of specific type is pressed at specific floor
func GetButton(button ButtonType, floor int) bool {
	a := read([4]byte{6, byte(button), byte(floor), 0})
	return toBool(a[1])
}

// GetFloor returns the current floor
func GetFloor() int {
	a := read([4]byte{7, 0, 0, 0})
	if a[1] != 0 {
		return int(a[2])
	} else {
		return -1
	}
}

// GetStop returns true if stop-button is pressed
func GetStop() bool {
	a := read([4]byte{8, 0, 0, 0})
	return toBool(a[1])
}

// GetObstruction returns true if obstruction switch is on
func GetObstruction() bool {
	a := read([4]byte{9, 0, 0, 0})
	return toBool(a[1])
}

// read states from hardware or simulator
func read(in [4]byte) [4]byte {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	var out [4]byte
	_, err = _conn.Read(out[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}

	return out
}

// write states from hardware or simulator
func write(in [4]byte) {
	_mtx.Lock()
	defer _mtx.Unlock()

	_, err := _conn.Write(in[:])
	if err != nil {
		panic("Lost connection to Elevator Server")
	}
}

// toByte returns 1 for true, and 0 for false
func toByte(a bool) byte {
	var b byte = 0
	if a {
		b = 1
	}
	return b
}

// toBool returns true for 1, and false for 0
func toBool(a byte) bool {
	var b bool = false
	if a != 0 {
		b = true
	}
	return b
}

func ToStringMotorDirection(direction MotorDirection) string {
	return _motorDirectionToString[direction]
}
