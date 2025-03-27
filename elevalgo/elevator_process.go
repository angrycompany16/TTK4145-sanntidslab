package elevalgo

import (
	"sanntidslab/elevio"
	"time"
)

type commandType int

const (
	setMotorDirection commandType = iota
	doorRequest
	setFloorIndicator
)

type elevatorCommands struct {
	_type commandType
	value any
}

// Runs a simple state machine for the elevator, which interfaces with the driver. It
// also manages a timer that panics if the motor loses power.
func RunElevator(
	floorChan <-chan int,
	orderChan <-chan elevio.ButtonEvent,
	doorCloseChan <-chan int,

	doorRequestChan chan<- int,
	nodeElevatorStateChan chan<- Elevator,
	lightsElevatorStateChan chan<- Elevator,
	resetMotorTimerChan chan<- int,
	stopMotorTimerChan chan<- int,

	config Config,
) {
	var commands []elevatorCommands

	elevatorInstance := NewUninitializedElevator(config)
	elevatorInstance, commands = initBetweenFloors(elevatorInstance)

	executeCommands(commands, doorRequestChan, resetMotorTimerChan, stopMotorTimerChan)

	for {
		select {
		case order := <-orderChan:
			elevatorInstance, commands = requestButtonPressed(elevatorInstance, order.Floor, order.Button)

			nodeElevatorStateChan <- elevatorInstance
			lightsElevatorStateChan <- elevatorInstance
		case floor := <-floorChan:
			elevatorInstance, commands = onFloorArrival(elevatorInstance, floor)

			nodeElevatorStateChan <- elevatorInstance
			lightsElevatorStateChan <- elevatorInstance
		case <-doorCloseChan:
			elevatorInstance, commands = onDoorClose(elevatorInstance)

			nodeElevatorStateChan <- elevatorInstance
			lightsElevatorStateChan <- elevatorInstance
		default:
			time.Sleep(time.Millisecond * 10)
			continue
		}
		executeCommands(commands, doorRequestChan, resetMotorTimerChan, stopMotorTimerChan)
	}
}

func executeCommands(
	commands []elevatorCommands,
	doorRequestChan chan<- int,
	resetMotorTimerChan chan<- int,
	stopMotorTimerChan chan<- int,
) {
	for _, command := range commands {
		switch command._type {
		case setMotorDirection:
			direction := command.value.(elevio.MotorDirection)
			elevio.SetMotorDirection(direction)
			if direction != elevio.MD_Stop {
				resetMotorTimerChan <- 1
			} else {
				stopMotorTimerChan <- 1
			}
		case setFloorIndicator:
			elevio.SetFloorIndicator(command.value.(int))
			stopMotorTimerChan <- 1
		case doorRequest:
			doorRequestChan <- 1
		}
	}
}
