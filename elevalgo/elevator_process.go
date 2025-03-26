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

// Runs an elevator that maintains a finite state machine and communicates with hardware
func RunElevator(
	floorChan <-chan int,
	orderChan <-chan elevio.ButtonEvent,
	doorCloseChan <-chan int,

	doorRequestChan chan<- int,
	nodeElevatorStateChan chan<- Elevator,
	lightsElevatorStateChan chan<- Elevator,
	resetMotorTimerChan chan<- int,
	stopMotorTimerChan chan<- int,
) {
	var commands []elevatorCommands
	var newElevator Elevator

	for {
		select {
		case order := <-orderChan:
			newElevator, commands = requestButtonPressed(elevator, order.Floor, order.Button)
			elevator = newElevator

			nodeElevatorStateChan <- newElevator
			lightsElevatorStateChan <- newElevator
		case floor := <-floorChan:
			newElevator, commands = onFloorArrival(elevator, floor)

			elevator = newElevator

			nodeElevatorStateChan <- newElevator
			lightsElevatorStateChan <- newElevator
		case <-doorCloseChan:
			newElevator, commands = onDoorClose(elevator)
			elevator = newElevator

			nodeElevatorStateChan <- newElevator
			lightsElevatorStateChan <- newElevator
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
		case doorRequest:
			doorRequestChan <- 1
		}
	}
}
