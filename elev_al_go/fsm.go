package elevalgo

import (
	"fmt"
	"runtime"
	"time"

	"github.com/angrycompany16/driver-go/elevio"
)

var (
	elevator Elevator
	timer    = MakeTimer(time.Second * 3)
)

func MakeFsm() Elevator {
	elevator = MakeUninitializedelevator()
	elevator.config.clearRequestVariation = CV_InDirn
	return elevator
}

func SetAllLights(elevator Elevator) {
	for floor := 0; floor < NUM_FLOORS; floor++ {
		for btn := 0; btn < NUM_BUTTONS; btn++ {
			elevio.SetButtonLamp(elevio.BT_HallDown, btn, elevator.requests[floor][btn])
			elevio.SetButtonLamp(elevio.BT_HallUp, btn, elevator.requests[floor][btn])
		}
	}
}

func FsmOnInitBetweenFloors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.direction = DIR_DOWN
	elevator.behaviour = BEHAVIOUR_MOVING
}

func FsmOnRequestButtonPress(btn_floor int, btn_type Button) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	fmt.Printf("\n\n%s(%d, %s)\n", frame.Function, btn_floor, ElevioButtonToString(btn_type))
	elevator.print()

	switch elevator.behaviour {
	case BEHAVIOUR_DOOR_OPEN:
		if elevator.RequestsShouldClearImmediately(btn_floor, btn_type) {
			timer.Start()
		} else {
			elevator.requests[btn_floor][btn_type] = true
		}
	case BEHAVIOUR_MOVING:
		elevator.requests[btn_floor][btn_type] = true
	case BEHAVIOUR_IDLE:
		elevator.requests[btn_floor][btn_type] = true
		pair := elevator.RequestsChooseDirection()
		elevator.direction = pair.dir
		elevator.behaviour = pair.behaviour
		switch pair.behaviour {
		case BEHAVIOUR_DOOR_OPEN:
			elevio.SetDoorOpenLamp(true)
			timer.Start()
			elevator = RequestsClearAtCurrentFloor(elevator)
		case BEHAVIOUR_MOVING:
			elevio.SetMotorDirection(elevio.MotorDirection(elevator.direction))
		}
	}

	SetAllLights(elevator)

	fmt.Printf("\nNew state:\n")
	elevator.print()
}

func FsmOnFloorArrival(newFloor int) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	fmt.Printf("\n\n%s(%d)\n", frame.Function, newFloor)
	elevator.print()

	elevator.floor = newFloor

	elevio.SetFloorIndicator(elevator.floor)

	switch elevator.behaviour {
	case BEHAVIOUR_MOVING:
		if elevator.RequestsShouldStop() {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			elevator = RequestsClearAtCurrentFloor(elevator)
			timer.Start()
			SetAllLights(elevator)
			elevator.behaviour = BEHAVIOUR_DOOR_OPEN
		}
	}

	fmt.Printf("\nNew state:\n")
	elevator.print()
}

func FsmOnDoorTimeout() {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	fmt.Printf("\n\n%s()\n", frame.Function)
	elevator.print()

	switch elevator.behaviour {
	case BEHAVIOUR_DOOR_OPEN:
		pair := elevator.RequestsChooseDirection()
		elevator.direction = pair.dir
		elevator.behaviour = pair.behaviour

		switch elevator.behaviour {
		case BEHAVIOUR_DOOR_OPEN:
			timer.Start()
			elevator = RequestsClearAtCurrentFloor(elevator)
			SetAllLights(elevator)
		case BEHAVIOUR_IDLE:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MotorDirection(elevator.direction))
		}
	}

	fmt.Printf("\nNew state:\n")
	elevator.print()
}
