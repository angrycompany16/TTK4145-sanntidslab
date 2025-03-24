package elevalgo

import (
	"fmt"
	"log"
	"sanntidslab/elevio"
)

const (
	NumFloors      = 4
	NumCabButtons  = 1
	NumHallButtons = 2
	NumButtons     = NumCabButtons + NumHallButtons
)

type elevatorBehaviour int

const (
	idle elevatorBehaviour = iota
	doorOpen
	moving
)

type direction int

const (
	down direction = iota - 1
	stop
	up
)

type Elevator struct {
	Floor     int
	direction direction
	Requests  [NumFloors][NumButtons]bool
	Behaviour elevatorBehaviour
	config    config
}

type dirBehaviourPair struct {
	dir       direction
	behaviour elevatorBehaviour
}

func dirToString(d direction) string {
	switch d {
	case up:
		return "D_Up"
	case down:
		return "D_Down"
	case stop:
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

func behaviourToString(behaviour elevatorBehaviour) string {
	switch behaviour {
	case idle:
		return "EB_Idle"
	case doorOpen:
		return "EB_DoorOpen"
	case moving:
		return "EB_Moving"
	default:
		return "EB_UNDEFINED"
	}
}

func (e *Elevator) print() {
	fmt.Println("  +--------------------+")
	fmt.Printf("  |floor = %-2d          |\n", e.Floor)
	fmt.Printf("  |dirn  = %-12.12s|\n", dirToString(e.direction))
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

func MakeUninitializedelevator() Elevator {
	config, err := loadConfig()
	if err != nil {
		log.Fatal("Failed to initialize elevator from .yaml file")
	}

	return Elevator{
		Floor:     -1,
		direction: stop,
		Behaviour: idle,
		config:    config,
	}
}

func ExtractCabCalls(elevator Elevator) (calls [NumFloors]bool) {
	for i := range NumFloors {
		// TODO: Make it general
		calls[i] = elevator.Requests[i][2]
	}
	return
}
