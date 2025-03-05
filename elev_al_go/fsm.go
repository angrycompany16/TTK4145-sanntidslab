package elevalgo

import (
	"fmt"
	"runtime"
	timer "sanntidslab/elev_al_go/timer"
	"time"

	"github.com/angrycompany16/driver-go/elevio"
)

var (
	ThisElevator Elevator
)

func InitFsm() {
	ThisElevator = MakeUninitializedelevator()
	initBetweenFloors()
}

func setAllLights(elevator Elevator) {
	for floor := 0; floor < NumFloors; floor++ {
		for btn := 0; btn < NumButtons; btn++ {
			if floor == 4 {
				fmt.Println("floor ", floor)
				fmt.Println("button ", btn)
			}
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, elevator.Requests[floor][btn])
		}
	}
}

func initBetweenFloors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	ThisElevator.direction = down
	ThisElevator.behaviour = moving
}

func RequestButtonPressed(buttonFloor int, buttonType elevio.ButtonType) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	fmt.Printf("\n\n%s(%d, %s)\n", frame.Function, buttonFloor, buttonToString(buttonType))
	ThisElevator.print()

	switch ThisElevator.behaviour {
	case doorOpen:
		if ThisElevator.shouldClearImmediately(buttonFloor, buttonType) {
			timer.StartTimer()
		} else {
			ThisElevator.Requests[buttonFloor][buttonType] = true
		}
	case moving:
		ThisElevator.Requests[buttonFloor][buttonType] = true
	case idle:
		ThisElevator.Requests[buttonFloor][buttonType] = true
		pair := ThisElevator.chooseDirection()
		ThisElevator.direction = pair.dir
		ThisElevator.behaviour = pair.behaviour
		switch pair.behaviour {
		case doorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.StartTimer()
			ThisElevator = clearAtCurrentFloor(ThisElevator)
		case moving:
			elevio.SetMotorDirection(elevio.MotorDirection(ThisElevator.direction))
		}
	}

	setAllLights(ThisElevator)

	fmt.Printf("\nNew state:\n")
	ThisElevator.print()
}

func OnFloorArrival(newFloor int) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	fmt.Printf("\n\n%s(%d)\n", frame.Function, newFloor)
	ThisElevator.print()

	ThisElevator.floor = newFloor

	elevio.SetFloorIndicator(ThisElevator.floor)

	switch ThisElevator.behaviour {
	case moving:
		if ThisElevator.shouldStop() {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			ThisElevator = clearAtCurrentFloor(ThisElevator)
			timer.StartTimer()
			setAllLights(ThisElevator)
			ThisElevator.behaviour = doorOpen
		}
	}

	fmt.Printf("\nNew state:\n")
	ThisElevator.print()
}

func OnDoorTimeout() {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	fmt.Printf("\n\n%s()\n", frame.Function)
	ThisElevator.print()

	switch ThisElevator.behaviour {
	case doorOpen:
		pair := ThisElevator.chooseDirection()
		ThisElevator.direction = pair.dir
		ThisElevator.behaviour = pair.behaviour

		switch ThisElevator.behaviour {
		case doorOpen:
			timer.StartTimer()
			ThisElevator = clearAtCurrentFloor(ThisElevator)
			setAllLights(ThisElevator)
		case moving, idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MotorDirection(ThisElevator.direction))
		}
	}

	fmt.Printf("\nNew state:\n")
	ThisElevator.print()
}

func DoorObstructed() {
	if ThisElevator.behaviour == doorOpen {
		timer.StartTimer()
	}
}

func GetTimeout() time.Duration {
	return ThisElevator.config.DoorOpenDuration
}
