package elevalgo

import (
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"github.com/angrycompany16/driver-go/elevio"
	"github.com/go-yaml/yaml"
)

const (
	NumFloors  = 4
	NumButtons = 3
)

var ConfigPath = path.Join("elev_al_go", "elevator_config.yaml")

type elevatorBehaviour int

const (
	idle elevatorBehaviour = iota
	doorOpen
	moving
)

type clearRequestVariant int

const (
	// Assume everyone waiting for the elevator gets on the elevator, even if
	// they will be traveling in the "wrong" direction for a while
	clearAll clearRequestVariant = iota

	// Assume that only those that want to travel in the current direction
	// enter the elevator, and keep waiting outside otherwise
	clearSameDir
)

type direction int

const (
	down direction = iota - 1
	stop
	up
)

type Elevator struct {
	floor     int
	direction direction
	Requests  [NumFloors][NumButtons]bool
	behaviour elevatorBehaviour
	config    config
}

type config struct {
	ClearRequestVariant clearRequestVariant `yaml:"ClearRequestVariant"`
	DoorOpenDuration    time.Duration       `yaml:"DoorOpenDuration"`
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

func buttonToString(b elevio.ButtonType) string {
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
	fmt.Printf("  |floor = %-2d          |\n", e.floor)
	fmt.Printf("  |dirn  = %-12.12s|\n", dirToString(e.direction))
	fmt.Printf("  |behav = %-12.12s|\n", behaviourToString(e.behaviour))

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

func loadConfig() (config, error) {
	c := config{}
	file, err := os.Open(ConfigPath)
	if err != nil {
		fmt.Println("Error reading file")
		return c, err
	}
	defer file.Close()

	err = yaml.NewDecoder(file).Decode(&c)
	if err != nil {
		fmt.Println("Error decoding file")
		return c, err
	}
	return c, nil
}

func MakeUninitializedelevator() Elevator {
	config, err := loadConfig()
	if err != nil {
		log.Fatal("Failed to initialize elevator from .yaml file")
	}

	return Elevator{
		floor:     -1,
		direction: stop,
		behaviour: idle,
		config:    config,
	}
}
