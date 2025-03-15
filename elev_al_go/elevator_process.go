package elevalgo

import (
	"os/exec"
	timer "sanntidslab/elev_al_go/timer"
	requests "sanntidslab/p2p/requests"

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
	orderChan <-chan requests.RequestInfo,
	/* output only channels */
	elevatorStateChan chan<- Elevator,
) {

	var commands []hardwareEffect
	var newElevator Elevator
	for {
		select {
		case requestInfo := <-orderChan:
			newElevator, commands = requestButtonPressed(elevator, requestInfo.Floor, requestInfo.ButtonType)
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
