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

func InitBetweenFloors() {
	elevio.SetMotorDirection(elevio.MD_Down)
	elevator.direction = down
	elevator.behaviour = moving
}

// TODO: Now we have no struct in this, but we still have a singleton (even worse, it's
// public...), and sure, defining these functions as methods on the struct means that
// they can modify any value in the struct, but they can do that now as well???
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
			commands = append(commands, hardwareEffect{effect: setMotorDirection, value: newElevator.direction})
		}
	}
	return newElevator, commands
}

// 	pc := make([]uintptr, 15)
// 	n := runtime.Callers(2, pc)
// 	frames := runtime.CallersFrames(pc[:n])
// 	frame, _ := frames.Next()

// 	if !DisablePrinting {
// 		fmt.Printf("\n\n%s(%d, %s)\n", frame.Function, buttonFloor, buttonToString(buttonType))
// 		elevator.print()
// 	}

// 	switch elevator.behaviour {
// 	case doorOpen:
// 		if elevator.shouldClearImmediately(buttonFloor, buttonType) {
// 			timer.StartTimer()
// 		} else {
// 			elevator.Requests[buttonFloor][buttonType] = true
// 		}
// 	case moving:
// 		elevator.Requests[buttonFloor][buttonType] = true
// 	case idle:
// 		elevator.Requests[buttonFloor][buttonType] = true
// 		pair := elevator.chooseDirection()
// 		elevator.direction = pair.dir
// 		elevator.behaviour = pair.behaviour
// 		switch pair.behaviour {
// 		case doorOpen:
// 			elevio.SetDoorOpenLamp(true)
// 			timer.StartTimer()
// 			elevator = clearAtCurrentFloor(elevator)
// 		case moving:
// 			elevio.SetMotorDirection(elevio.MotorDirection(elevator.direction))
// 		}
// 	}

// 	// setAllLights(e)

// 	if !DisablePrinting {
// 		fmt.Printf("\nNew state:\n")
// 		elevator.print()
// 	}
// }

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
			commands = append(commands, hardwareEffect{effect: setMotorDirection, value: elevio.MD_Stop})
			commands = append(commands, hardwareEffect{effect: setDoorOpenLamp, value: true})
			newElevator = clearAtCurrentFloor(newElevator)
			commands = append(commands, hardwareEffect{effect: startTimer, value: nil})
			newElevator.behaviour = doorOpen
		}
	}
	return newElevator, commands
}

// 	pc := make([]uintptr, 15)
// 	n := runtime.Callers(2, pc)
// 	frames := runtime.CallersFrames(pc[:n])
// 	frame, _ := frames.Next()

// 	if !DisablePrinting {
// 		fmt.Printf("\n\n%s(%d)\n", frame.Function, newFloor)
// 		elevator.print()
// 	}
// 	elevator.floor = newFloor

// 	elevio.SetFloorIndicator(elevator.floor)

// 	switch elevator.behaviour {
// 	case moving:
// 		if elevator.shouldStop() {
// 			elevio.SetMotorDirection(elevio.MD_Stop)
// 			elevio.SetDoorOpenLamp(true)
// 			elevator = clearAtCurrentFloor(elevator)
// 			timer.StartTimer()
// 			// setAllLights(e)
// 			elevator.behaviour = doorOpen
// 		}
// 	}

// 	if !DisablePrinting {
// 		fmt.Printf("\nNew state:\n")
// 		elevator.print()
// 	}
// }

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

// 	pc := make([]uintptr, 15)
// 	n := runtime.Callers(2, pc)
// 	frames := runtime.CallersFrames(pc[:n])
// 	frame, _ := frames.Next()

// 	if !DisablePrinting {
// 		fmt.Printf("\n\n%s()\n", frame.Function)
// 		elevator.print()
// 	}
// 	switch elevator.behaviour {
// 	case doorOpen:
// 		pair := elevator.chooseDirection()
// 		elevator.direction = pair.dir
// 		elevator.behaviour = pair.behaviour

// 		switch elevator.behaviour {
// 		case doorOpen:
// 			timer.StartTimer()
// 			elevator = clearAtCurrentFloor(elevator)
// 			// setAllLights(e)
// 		case moving, idle:
// 			fmt.Println("Closing door")
// 			elevio.SetDoorOpenLamp(false)
// 			elevio.SetMotorDirection(elevio.MotorDirection(elevator.direction))
// 		}
// 	}

// 	if !DisablePrinting {
// 		fmt.Printf("\nNew state:\n")
// 		elevator.print()
// 	}
// }

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

// 	if !isObstructed {
// 		return
// 	}

// 	fmt.Println("obstr")

// 	if elevator.behaviour == doorOpen {
// 		timer.StartTimer()
// 	}
// }

func GetTimeout() time.Duration {
	return elevator.config.DoorOpenDuration
}

func GetRequestStatus(floor int, button int) bool {
	return elevator.Requests[floor][button]
}

func GetState() Elevator {
	return elevator
}
