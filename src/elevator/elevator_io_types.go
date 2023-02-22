package elevator

const N_FLOORS = 4
const N_BUTTONS = 3

type Dirn int

const (
	D_Down Dirn = -1
	D_Stop Dirn = 0
	D_Up   Dirn = 1
)

type Button int

const (
	B_HallUp Button = iota
	B_HallDown
	B_Cab
)

/* Some more stuff yolo
type ElevInputDevice struct {

}
*/
