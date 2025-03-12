package elevalgo

import (
	"fmt"
	"runtime"
	timer "sanntidslab/elev_al_go/timer"
	"time"

	"github.com/angrycompany16/driver-go/elevio"
)

const (
	DisablePrinting = true
)

var (
	elevator Elevator
)

// TODO: Rewrite to functional

func InitFsm() Elevator {
	return MakeUninitializedelevator()
}

func SetAllLights(elevator Elevator) {
	for floor := 0; floor < NumFloors; floor++ {
		for btn := 0; btn < NumButtons; btn++ {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, elevator.Requests[floor][btn])
		}
	}
}

func initBetweenFloors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.direction = down
	elevator.behaviour = moving
}

// TODO: Now we have no struct in this, but we still have a singleton (even worse, it's
// public...), and sure, defining these functions as methods on the struct means that
// they can modify any value in the struct, but they can do that now as well???
func RequestButtonPressed(buttonFloor int, buttonType elevio.ButtonType) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	if !DisablePrinting {
		fmt.Printf("\n\n%s(%d, %s)\n", frame.Function, buttonFloor, buttonToString(buttonType))
		elevator.print()
	}

	switch elevator.behaviour {
	case doorOpen:
		if elevator.shouldClearImmediately(buttonFloor, buttonType) {
			timer.StartTimer()
		} else {
			elevator.Requests[buttonFloor][buttonType] = true
		}
	case moving:
		elevator.Requests[buttonFloor][buttonType] = true
	case idle:
		elevator.Requests[buttonFloor][buttonType] = true
		pair := elevator.chooseDirection()
		elevator.direction = pair.dir
		elevator.behaviour = pair.behaviour
		switch pair.behaviour {
		case doorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.StartTimer()
			elevator = clearAtCurrentFloor(elevator)
		case moving:
			elevio.SetMotorDirection(elevio.MotorDirection(elevator.direction))
		}
	}

	// setAllLights(e)

	if !DisablePrinting {
		fmt.Printf("\nNew state:\n")
		elevator.print()
	}
}

func OnFloorArrival(newFloor int) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	if !DisablePrinting {
		fmt.Printf("\n\n%s(%d)\n", frame.Function, newFloor)
		elevator.print()
	}
	elevator.floor = newFloor

	elevio.SetFloorIndicator(elevator.floor)

	switch elevator.behaviour {
	case moving:
		if elevator.shouldStop() {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			elevator = clearAtCurrentFloor(elevator)
			timer.StartTimer()
			// setAllLights(e)
			elevator.behaviour = doorOpen
		}
	}

	if !DisablePrinting {
		fmt.Printf("\nNew state:\n")
		elevator.print()
	}
}

func OnDoorTimeout() {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	if !DisablePrinting {
		fmt.Printf("\n\n%s()\n", frame.Function)
		elevator.print()
	}
	switch elevator.behaviour {
	case doorOpen:
		pair := elevator.chooseDirection()
		elevator.direction = pair.dir
		elevator.behaviour = pair.behaviour

		switch elevator.behaviour {
		case doorOpen:
			timer.StartTimer()
			elevator = clearAtCurrentFloor(elevator)
			// setAllLights(e)
		case moving, idle:
			fmt.Println("Closing door")
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MotorDirection(elevator.direction))
		}
	}

	if !DisablePrinting {
		fmt.Printf("\nNew state:\n")
		elevator.print()
	}
}

func DoorObstructed(isObstructed bool) {
	if !isObstructed {
		return
	}

	fmt.Println("obstr")

	if elevator.behaviour == doorOpen {
		timer.StartTimer()
	}
}

func GetTimeout() time.Duration {
	return elevator.config.DoorOpenDuration
}

func GetRequestStatus(floor int, button int) bool {
	return elevator.Requests[floor][button]
}

func GetState() Elevator {
	return elevator
}
