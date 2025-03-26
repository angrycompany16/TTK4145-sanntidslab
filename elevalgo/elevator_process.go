package elevalgo

import (
	"sanntidslab/elevio"
	"time"
)

type driverType int

const (
	setMotorDirection driverType = iota
	doorRequest
	setFloorIndicator
)

type hardwareEffect struct {
	effect driverType
	value  interface{}
}

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
	var commands []hardwareEffect
	var newElevator Elevator

	for {
		select {
		case order := <-orderChan:
			newElevator, commands = requestButtonPressed(elevator, order.Floor, order.Button)
			elevator = newElevator

			// Twice, one for lights and one for network
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
	commands []hardwareEffect,
	doorRequestChan chan<- int,
	resetMotorTimerChan chan<- int,
	stopMotorTimerChan chan<- int,
) {
	for _, command := range commands {
		switch command.effect {
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
