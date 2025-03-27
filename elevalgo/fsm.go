package elevalgo

import (
	"fmt"
	"runtime"
	"sanntidslab/elevio"
	"time"
)

const (
	DisablePrinting = true // Only for debugging
)

var (
	elevator Elevator
)

func InitFsm() {
	elevator = MakeUninitializedelevator()
}

func InitBetweenFloors() (Elevator, time.Duration) {
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.direction = down
	elevator.Behaviour = moving
	return elevator, elevator.config.DoorOpenDuration
}

func requestButtonPressed(e Elevator, buttonFloor int, buttonType elevio.ButtonType) (newElevator Elevator, commands []elevatorCommands) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	if !DisablePrinting {
		fmt.Printf("\n\n%s(%d, %s)\n", frame.Function, buttonFloor, ButtonToString(buttonType))
		elevator.print()
	}

	newElevator = e
	switch newElevator.Behaviour {
	case doorOpen:
		if newElevator.shouldClearImmediately(buttonFloor, buttonType) {
			commands = append(commands, elevatorCommands{_type: doorRequest, value: nil})
		} else {
			newElevator.Requests[buttonFloor][buttonType] = true
		}
	case moving:
		newElevator.Requests[buttonFloor][buttonType] = true
	case idle:
		newElevator.Requests[buttonFloor][buttonType] = true
		pair := newElevator.chooseDirection()
		newElevator.direction = pair.dir
		newElevator.Behaviour = pair.behaviour
		switch pair.behaviour {
		case doorOpen:
			commands = append(commands, elevatorCommands{_type: doorRequest, value: true})
			newElevator = clearAtCurrentFloor(newElevator)
		case moving:
			commands = append(commands, elevatorCommands{_type: setMotorDirection, value: elevio.MotorDirection(newElevator.direction)})
		}
	}
	return
}

func onFloorArrival(e Elevator, newFloor int) (newElevator Elevator, commands []elevatorCommands) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	if !DisablePrinting {
		fmt.Printf("\n\n%s(%d)\n", frame.Function, newFloor)
		elevator.print()
	}

	newElevator = e
	newElevator.Floor = newFloor

	commands = append(commands, elevatorCommands{_type: setFloorIndicator, value: newFloor})

	switch newElevator.Behaviour {
	case moving:
		if newElevator.shouldStop() {
			commands = append(commands, elevatorCommands{_type: setMotorDirection, value: elevio.MotorDirection(elevio.MD_Stop)})
			commands = append(commands, elevatorCommands{_type: doorRequest, value: true})
			newElevator = clearAtCurrentFloor(newElevator)
			newElevator.Behaviour = doorOpen
		}
	}
	return
}

func onDoorClose(e Elevator) (newElevator Elevator, commands []elevatorCommands) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	if !DisablePrinting {
		fmt.Printf("\n\n%s()\n", frame.Function)
		elevator.print()
	}

	newElevator = e

	switch newElevator.Behaviour {
	case doorOpen:
		pair := newElevator.chooseDirection()
		newElevator.direction = pair.dir
		newElevator.Behaviour = pair.behaviour

		switch newElevator.Behaviour {
		case doorOpen:
			commands = append(commands, elevatorCommands{_type: doorRequest, value: nil})
			newElevator = clearAtCurrentFloor(newElevator)

		case moving, idle:
			commands = append(commands, elevatorCommands{_type: setMotorDirection, value: elevio.MotorDirection(newElevator.direction)})
		}
	}
	return
}
