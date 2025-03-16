package elevalgo

import (
	"fmt"
	"runtime"
	"time"

	"github.com/angrycompany16/driver-go/elevio"
)

const (
	DisablePrinting = true
)

var (
	elevator Elevator
)

func InitFsm() {
	elevator = MakeUninitializedelevator()
}

func SetAllLights(elevator Elevator) {
	for floor := range NumFloors {
		for btn := range NumButtons {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, elevator.Requests[floor][btn])
		}
	}
}

func InitBetweenFloors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.direction = down
	elevator.behaviour = moving
}

func requestButtonPressed(e Elevator, buttonFloor int, buttonType elevio.ButtonType) (newElevator Elevator, commands []hardwareEffect) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	if !DisablePrinting {
		fmt.Printf("\n\n%s(%d, %s)\n", frame.Function, buttonFloor, buttonToString(buttonType))
		elevator.print()
	}

	newElevator = e
	switch newElevator.behaviour {
	case doorOpen:
		if newElevator.shouldClearImmediately(buttonFloor, buttonType) {
			commands = append(commands, hardwareEffect{effect: startTimer, value: nil})
		} else {
			newElevator.Requests[buttonFloor][buttonType] = true
		}
	case moving:
		newElevator.Requests[buttonFloor][buttonType] = true
	case idle:
		newElevator.Requests[buttonFloor][buttonType] = true
		pair := newElevator.chooseDirection()
		newElevator.direction = pair.dir
		newElevator.behaviour = pair.behaviour
		switch pair.behaviour {
		case doorOpen:
			commands = append(commands, hardwareEffect{effect: setDoorOpenLamp, value: true})
			commands = append(commands, hardwareEffect{effect: startTimer, value: nil})
			newElevator = clearAtCurrentFloor(newElevator)
		case moving:
			commands = append(commands, hardwareEffect{effect: setMotorDirection, value: elevio.MotorDirection(newElevator.direction)})
		}
	}
	return newElevator, commands
}

func onFloorArrival(e Elevator, newFloor int) (newElevator Elevator, commands []hardwareEffect) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	if !DisablePrinting {
		fmt.Printf("\n\n%s(%d)\n", frame.Function, newFloor)
		elevator.print()
	}

	newElevator = e
	newElevator.floor = newFloor

	commands = append(commands, hardwareEffect{effect: setFloorIndicator, value: newFloor})

	switch newElevator.behaviour {
	case moving:
		if newElevator.shouldStop() {
			commands = append(commands, hardwareEffect{effect: setMotorDirection, value: elevio.MotorDirection(elevio.MD_Stop)})
			commands = append(commands, hardwareEffect{effect: setDoorOpenLamp, value: true})
			newElevator = clearAtCurrentFloor(newElevator)
			commands = append(commands, hardwareEffect{effect: startTimer, value: nil})
			newElevator.behaviour = doorOpen
		}
	}
	return newElevator, commands
}

func onDoorTimeout(e Elevator) (newElevator Elevator, commands []hardwareEffect) {
	pc := make([]uintptr, 15)
	n := runtime.Callers(2, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()

	if !DisablePrinting {
		fmt.Printf("\n\n%s()\n", frame.Function)
		elevator.print()
	}

	newElevator = e

	switch newElevator.behaviour {
	case doorOpen:
		pair := newElevator.chooseDirection()
		newElevator.direction = pair.dir
		newElevator.behaviour = pair.behaviour

		switch newElevator.behaviour {
		case doorOpen:
			commands = append(commands, hardwareEffect{effect: startTimer, value: nil})
			newElevator = clearAtCurrentFloor(newElevator)

		case moving, idle:
			commands = append(commands, hardwareEffect{effect: setDoorOpenLamp, value: false})
			commands = append(commands, hardwareEffect{effect: setMotorDirection, value: elevio.MotorDirection(newElevator.direction)})
		}
	}
	return newElevator, commands
}

func doorObstructed(e Elevator, isObstructed bool) (commands []hardwareEffect) {
	if !isObstructed {
		return commands
	}
	fmt.Println("Obstructed")

	if e.behaviour == doorOpen {
		commands = append(commands, hardwareEffect{effect: startTimer, value: nil})
	}
	return commands
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
