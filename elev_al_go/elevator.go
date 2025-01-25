package elevalgo

import (
	"fmt"
	"time"
)

type ElevatorBehaviour int

const (
	BEHAVIOUR_IDLE = iota
	BEHAVIOUR_DOOR_OPEN
	BEHAVIOUR_MOVING
)

type ClearRequestVariant int

const (
	// Assume everyone waiting for the elevator gets on the elevator, even if
	// they will be traveling in the "wrong" direction for a while
	CV_All = iota

	// Assume that only those that want to travel in the current direction
	// enter the elevator, and keep waiting outside otherwise
	CV_InDirn
)

type Elevator struct {
	floor     int
	direction Dir
	requests  [NUM_FLOORS][NUM_BUTTONS]bool
	behaviour ElevatorBehaviour
	config    config
}

type config struct {
	clearRequestVariation ClearRequestVariant
	doorOpenDuration      time.Duration
}

type DirBehaviourPair struct {
	dir       Dir
	behaviour ElevatorBehaviour
}

const (
	NUM_FLOORS  = 4
	NUM_BUTTONS = 3
)

type Dir int

const (
	DIR_DOWN Dir = iota - 1
	DIR_STOP
	DIR_UP
)

type Button int

const (
	BTN_HALLUP Button = iota
	BTN_HALLDOWN
	BTN_HALLCAB
)

func ElevioDirToString(d Dir) string {
	switch d {
	case DIR_UP:
		return "D_Up"
	case DIR_DOWN:
		return "D_Down"
	case DIR_STOP:
		return "D_Stop"
	default:
		return "D_UNDEFINED"
	}
}

func ElevioButtonToString(b Button) string {
	switch b {
	case BTN_HALLUP:
		return "B_HallUp"
	case BTN_HALLDOWN:
		return "B_HallDown"
	case BTN_HALLCAB:
		return "B_Cab"
	default:
		return "B_UNDEFINED"
	}
}

func EbToString(behaviour ElevatorBehaviour) string {
	switch behaviour {
	case BEHAVIOUR_IDLE:
		return "EB_Idle"
	case BEHAVIOUR_DOOR_OPEN:
		return "EB_DoorOpen"
	case BEHAVIOUR_MOVING:
		return "EB_Moving"
	default:
		return "EB_UNDEFINED"
	}
}

func (e *Elevator) print() {
	fmt.Println("  +--------------------+")
	fmt.Printf("  |floor = %-2d          |\n", e.floor)
	fmt.Printf("  |dirn  = %-12.12s|\n", ElevioDirToString(e.direction))
	fmt.Printf("  |behav = %-12.12s|\n", EbToString(e.behaviour))

	fmt.Println("  +--------------------+")
	fmt.Println("  |  | up  | dn  | cab |")
	for f := NUM_FLOORS - 1; f >= 0; f-- {
		fmt.Printf("  | %d", f)
		for btn := 0; btn < NUM_BUTTONS; btn++ {
			if (f == NUM_FLOORS-1 && btn == int(BTN_HALLUP)) || (f == 0 && btn == int(BTN_HALLDOWN)) {
				fmt.Print("|     ")
			} else {
				if e.requests[f][btn] {
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
	return Elevator{
		floor:     -1,
		direction: DIR_STOP,
		behaviour: BEHAVIOUR_IDLE,
		config: config{
			clearRequestVariation: CV_All,
			doorOpenDuration:      3.0,
		},
	}
}
