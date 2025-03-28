package elevalgo

import (
	"fmt"
	"sanntidslab/elevio"
)

type ElevatorBehaviour int

const (
	Idle ElevatorBehaviour = iota
	DoorOpen
	Moving
)

type Direction int

const (
	Down Direction = iota - 1
	Stop
	Up
)

type Elevator struct {
	Floor     int
	Direction Direction
	Requests  [NumFloors][NumButtons]bool
	Behaviour ElevatorBehaviour
	config    Config
}

type dirBehaviourPair struct {
	dir       Direction
	behaviour ElevatorBehaviour
}

func dirToString(d Direction) string {
	switch d {
	case Up:
		return "D_Up"
	case Down:
		return "D_Down"
	case Stop:
		return "D_Stop"
	default:
		return "D_UNDEFINED"
	}
}

func ButtonToString(b elevio.ButtonType) string {
	switch b {
	case elevio.BT_HallUp:
		return "B_HallUp"
	case elevio.BT_HallDown:
		return "B_HallDown"
	case elevio.BT_Cab:
		return "B_Cab"
	default:
		return "B_UNDEFINED"
	}
}

func behaviourToString(behaviour ElevatorBehaviour) string {
	switch behaviour {
	case Idle:
		return "EB_Idle"
	case DoorOpen:
		return "EB_DoorOpen"
	case Moving:
		return "EB_Moving"
	default:
		return "EB_UNDEFINED"
	}
}

func (e *Elevator) print() {
	fmt.Println("  +--------------------+")
	fmt.Printf("  |floor = %-2d          |\n", e.Floor)
	fmt.Printf("  |dirn  = %-12.12s|\n", dirToString(e.Direction))
	fmt.Printf("  |behav = %-12.12s|\n", behaviourToString(e.Behaviour))

	fmt.Println("  +--------------------+")
	fmt.Println("  |  | up  | dn  | cab |")
	for f := NumFloors - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		for btn := 0; btn < NumButtons; btn++ {
			if (f == NumFloors-1 && btn == int(elevio.BT_HallUp)) || (f == 0 && btn == int(elevio.BT_HallDown)) {
				fmt.Print("|     ")
			} else {
				if e.Requests[f][btn] {
					fmt.Print("|  #  ")
				} else {
					fmt.Print("|  -  ")
				}
			}
		}
		fmt.Println("|")
	}
	fmt.Println("  +--------------------+")
}

func NewUninitializedElevator(config Config) Elevator {
	return Elevator{
		Floor:     -1,
		Direction: Stop,
		Behaviour: Idle,
		config:    config,
	}
}

func ExtractCabCalls(elevator Elevator) (calls [NumFloors]bool) {
	for i := range NumFloors {
		calls[i] = elevator.Requests[i][2]
	}
	return
}
