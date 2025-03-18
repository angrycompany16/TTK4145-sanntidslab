package elevalgo

import (
	"github.com/angrycompany16/driver-go/elevio"
)

type driverType int

const (
	setMotorDirection driverType = iota
	setDoorOpenLamp
	setFloorIndicator
	startTimer
)

type hardwareEffect struct {
	effect driverType
	value  interface{}
}

func RunElevator(
	/* input only channels */
	floorChan <-chan int,
	obstructionChan <-chan bool,
	orderChan <-chan elevio.ButtonEvent,
	doorTimeoutChan <-chan int,
	peerStatesChan <-chan []Elevator,

	/* output only channels */
	elevatorStateChan chan<- Elevator,
	startDoorTimerChan chan<- int,
	startObstructionTimerChan chan<- int,
	startMotorTimerChan chan<- int,
) {
	var commands []hardwareEffect
	var newElevator Elevator

	for {
		select {
		case requestInfo := <-orderChan:
			newElevator, commands = requestButtonPressed(elevator, requestInfo.Floor, requestInfo.Button)
			elevator = newElevator
		case floor := <-floorChan:
			newElevator, commands = onFloorArrival(elevator, floor)
			elevator = newElevator
		case obstructionEvent := <-obstructionChan:
			commands = doorObstructed(elevator, obstructionEvent)
			if !obstructionEvent {
				startObstructionTimerChan <- 1
			}
		case <-doorTimeoutChan:
			newElevator, commands = onDoorTimeout(elevator)
			elevator = newElevator
		case peerStates := <-peerStatesChan:
			lightsState := MergeHallLights(elevator, append(peerStates, elevator))
			SetLights(lightsState)
		default:
			elevatorStateChan <- GetState()
			continue
		}
		executeCommands(commands, startDoorTimerChan, startMotorTimerChan)
	}
}

func executeCommands(
	commands []hardwareEffect,
	startDoorTimerChan chan<- int,
	startMotorTimerChan chan<- int,
) {
	for _, command := range commands {
		switch command.effect {
		case setMotorDirection:
			direction := command.value.(elevio.MotorDirection)
			elevio.SetMotorDirection(direction)
			if direction == elevio.MD_Up || direction == elevio.MD_Down {
				startMotorTimerChan <- 1
			}
		case setDoorOpenLamp:
			elevio.SetDoorOpenLamp(command.value.(bool))
		case setFloorIndicator:
			elevio.SetFloorIndicator(command.value.(int))
		case startTimer:
			startDoorTimerChan <- 1
		}
	}
}
