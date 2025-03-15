package elevalgo

import (
	timer "sanntidslab/elev_al_go/timer"

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

func ElevatorProcess(
	/* input only channels */
	floorChan <-chan int,
	obstructionChan <-chan bool,
	orderChan <-chan elevio.ButtonEvent,
	/* output only channels */
	elevatorStateChan chan<- Elevator,
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

		case <-timer.TimeoutChan:
			timer.StopTimer()
			newElevator, commands = onDoorTimeout(elevator)

			elevator = newElevator
		default:
			elevatorStateChan <- GetState()
			timer.CheckTimeout()
			continue
		}
		executeCommands(commands)
	}
}

func executeCommands(commands []hardwareEffect) {
	for _, command := range commands {
		switch command.effect {
		case setMotorDirection:
			elevio.SetMotorDirection(command.value.(elevio.MotorDirection))
		case setDoorOpenLamp:
			elevio.SetDoorOpenLamp(command.value.(bool))
		case setFloorIndicator:
			elevio.SetFloorIndicator(command.value.(int))
		case startTimer:
			timer.StartTimer()
		}
	}
}
