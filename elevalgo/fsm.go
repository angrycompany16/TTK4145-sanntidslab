package elevalgo

import (
	"fmt"
	"runtime"
	"sanntidslab/elevio"
)

func initBetweenFloors(elevator Elevator) (newElevator Elevator, commands []elevatorCommands) {
	commands = append(commands, elevatorCommands{_type: setMotorDirection, value: elevio.MD_Down})
	newElevator = elevator
	newElevator.Direction = Down
	newElevator.Behaviour = moving
	return
}

func requestButtonPressed(elevator Elevator, buttonFloor int, buttonType elevio.ButtonType) (newElevator Elevator, commands []elevatorCommands) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	fmt.Printf("\n\n%s(%d, %s)\n", frame.Function, buttonFloor, ButtonToString(buttonType))
	elevator.print()

	newElevator = elevator
	switch newElevator.Behaviour {
	case doorOpen:
		if shouldClearImmediately(newElevator, buttonFloor, buttonType) {
			commands = append(commands, elevatorCommands{_type: doorRequest, value: nil})
		} else {
			newElevator.Requests[buttonFloor][buttonType] = true
		}
	case moving:
		newElevator.Requests[buttonFloor][buttonType] = true
	case idle:
		newElevator.Requests[buttonFloor][buttonType] = true
		pair := chooseDirection(newElevator)
		newElevator.Direction = pair.dir
		newElevator.Behaviour = pair.behaviour
		switch pair.behaviour {
		case doorOpen:
			commands = append(commands, elevatorCommands{_type: doorRequest, value: true})
			newElevator = clearAtCurrentFloor(newElevator)
		case moving:
			commands = append(commands, elevatorCommands{_type: setMotorDirection, value: elevio.MotorDirection(newElevator.Direction)})
		}
	}
	return
}

func onFloorArrival(elevator Elevator, newFloor int) (newElevator Elevator, commands []elevatorCommands) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	fmt.Printf("\n\n%s(%d)\n", frame.Function, newFloor)
	elevator.print()

	newElevator = elevator
	newElevator.Floor = newFloor

	commands = append(commands, elevatorCommands{_type: setFloorIndicator, value: newFloor})

	switch newElevator.Behaviour {
	case moving:
		if shouldStop(newElevator) {
			commands = append(commands, elevatorCommands{_type: setMotorDirection, value: elevio.MotorDirection(elevio.MD_Stop)})
			commands = append(commands, elevatorCommands{_type: doorRequest, value: true})
			newElevator = clearAtCurrentFloor(newElevator)
			newElevator.Behaviour = doorOpen
		}
	}
	return
}

func onDoorClose(elevator Elevator) (newElevator Elevator, commands []elevatorCommands) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	fmt.Printf("\n\n%s()\n", frame.Function)
	elevator.print()

	newElevator = elevator

	switch newElevator.Behaviour {
	case doorOpen:
		pair := chooseDirection(newElevator)
		newElevator.Direction = pair.dir
		newElevator.Behaviour = pair.behaviour

		switch newElevator.Behaviour {
		case doorOpen:
			commands = append(commands, elevatorCommands{_type: doorRequest, value: nil})
			newElevator = clearAtCurrentFloor(newElevator)

		case moving, idle:
			commands = append(commands, elevatorCommands{_type: setMotorDirection, value: elevio.MotorDirection(newElevator.Direction)})
		}
	}
	return
}
