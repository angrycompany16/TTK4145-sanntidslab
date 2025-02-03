package elevalgo

import (
	"fmt"
	"runtime"
	timer "sanntidslab/elev_al_go/timer"
	"time"

	"github.com/angrycompany16/driver-go/elevio"
)

var (
	runningElevator elevator
)

func InitFsm() {
	runningElevator = MakeUninitializedelevator()
	initBetweenFloors()
}

func setAllLights(elevator elevator) {
	for floor := 0; floor < NumFloors; floor++ {
		for btn := 0; btn < NumButtons; btn++ {
			if floor == 4 {
				fmt.Println("floor ", floor)
				fmt.Println("button ", btn)
			}
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, elevator.requests[floor][btn])
		}
	}
}

func initBetweenFloors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	runningElevator.direction = down
	runningElevator.behaviour = moving
}

func RequestButtonPressed(buttonFloor int, buttonType Button) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	fmt.Printf("\n\n%s(%d, %s)\n", frame.Function, buttonFloor, buttonToString(buttonType))
	runningElevator.print()

	switch runningElevator.behaviour {
	case doorOpen:
		if runningElevator.shouldClearImmediately(buttonFloor, buttonType) {
			timer.StartTimer()
		} else {
			runningElevator.requests[buttonFloor][buttonType] = true
		}
	case moving:
		runningElevator.requests[buttonFloor][buttonType] = true
	case idle:
		runningElevator.requests[buttonFloor][buttonType] = true
		pair := runningElevator.chooseDirection()
		runningElevator.direction = pair.dir
		runningElevator.behaviour = pair.behaviour
		switch pair.behaviour {
		case doorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.StartTimer()
			runningElevator = clearAtCurrentFloor(runningElevator)
		case moving:
			elevio.SetMotorDirection(elevio.MotorDirection(runningElevator.direction))
		}
	}

	setAllLights(runningElevator)

	fmt.Printf("\nNew state:\n")
	runningElevator.print()
}

func OnFloorArrival(newFloor int) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	fmt.Printf("\n\n%s(%d)\n", frame.Function, newFloor)
	runningElevator.print()

	runningElevator.floor = newFloor

	elevio.SetFloorIndicator(runningElevator.floor)

	switch runningElevator.behaviour {
	case moving:
		if runningElevator.shouldStop() {
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			runningElevator = clearAtCurrentFloor(runningElevator)
			timer.StartTimer()
			setAllLights(runningElevator)
			runningElevator.behaviour = doorOpen
		}
	}

	fmt.Printf("\nNew state:\n")
	runningElevator.print()
}

func OnDoorTimeout() {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	fmt.Printf("\n\n%s()\n", frame.Function)
	runningElevator.print()

	switch runningElevator.behaviour {
	case doorOpen:
		pair := runningElevator.chooseDirection()
		runningElevator.direction = pair.dir
		runningElevator.behaviour = pair.behaviour

		switch runningElevator.behaviour {
		case doorOpen:
			timer.StartTimer()
			runningElevator = clearAtCurrentFloor(runningElevator)
			setAllLights(runningElevator)
		case moving, idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elevio.MotorDirection(runningElevator.direction))
		}
	}

	fmt.Printf("\nNew state:\n")
	runningElevator.print()
}

func DoorObstructed() {
	if runningElevator.behaviour == doorOpen {
		timer.StartTimer()
	}
}

func GetTimeout() time.Duration {
	return runningElevator.config.DoorOpenDuration
}
