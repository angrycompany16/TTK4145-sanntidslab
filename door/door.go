package door

import (
	"fmt"
	"sanntidslab/elevio"
)

type DoorState int

const (
	Open DoorState = iota
	Closed
	Stuck
)

type doorCommand int

const (
	openDoor doorCommand = iota
	closeDoor
	resetTimer
	resetObstructionTimer
	stopObstructionTimer
)

func RunDoor(
	obstructionChan <-chan bool,
	doorTimeoutChan <-chan int,
	doorRequestChan <-chan int,

	doorCloseChan chan<- int,
	startDoorTimerChan chan<- int,
	resetObstructionTimerChan chan<- int,
	stopObstructionTimerChan chan<- int,

	startObstructed bool,
) {
	var doorInstanceState DoorState
	initCommands := []doorCommand{closeDoor}
	if startObstructed {
		doorInstanceState = Stuck
		initCommands = append(initCommands, resetObstructionTimer)
	} else {
		doorInstanceState = Closed
		initCommands = append(initCommands, stopObstructionTimer)
	}
	executeCommands(startDoorTimerChan, doorCloseChan, resetObstructionTimerChan, stopObstructionTimerChan, initCommands)

	for {
		var commands []doorCommand
		select {
		case obstructionEvent := <-obstructionChan:
			doorInstanceState, commands = onObstructionEvent(obstructionEvent, doorInstanceState)
		case <-doorTimeoutChan:
			doorInstanceState, commands = onDoorTimeout(doorInstanceState)
		case <-doorRequestChan:
			doorInstanceState, commands = onDoorRequest(doorInstanceState)
		}

		executeCommands(startDoorTimerChan, doorCloseChan, resetObstructionTimerChan, stopObstructionTimerChan, commands)
	}
}

func executeCommands(
	startDoorTimerChan chan<- int,
	doorCloseChan chan<- int,
	resetObstructionTimerChan chan<- int,
	stopObstructionTimerChan chan<- int,
	commands []doorCommand,
) {
	for _, command := range commands {
		switch command {
		case openDoor:
			elevio.SetDoorOpenLamp(true)
		case closeDoor:
			doorCloseChan <- 1
			elevio.SetDoorOpenLamp(false)
		case resetTimer:
			startDoorTimerChan <- 1
		case resetObstructionTimer:
			resetObstructionTimerChan <- 1
		case stopObstructionTimer:
			stopObstructionTimerChan <- 1
		}
	}
}

func onObstructionEvent(obstructionEvent bool, state DoorState) (newState DoorState, commands []doorCommand) {
	if obstructionEvent {
		commands = append(commands, resetObstructionTimer)
	} else {
		commands = append(commands, stopObstructionTimer)
	}

	switch state {
	case Open:
		newState = Stuck
	case Closed:
		newState = Stuck
	case Stuck:
		if !obstructionEvent {
			fmt.Println("Obstruction freed")
			newState = Open
			commands = append(commands, resetTimer)
			return
		}
		newState = Stuck
	}
	return
}

func onDoorTimeout(state DoorState) (DoorState, []doorCommand) {
	switch state {
	case Open:
		return Closed, []doorCommand{closeDoor}
	case Closed:
	case Stuck:
		return state, []doorCommand{resetTimer}
	}
	return state, nil
}

func onDoorRequest(state DoorState) (DoorState, []doorCommand) {
	switch state {
	case Open:
		return state, []doorCommand{resetTimer}
	case Closed:
		return Open, []doorCommand{resetTimer, openDoor}
	case Stuck:
		return Stuck, []doorCommand{resetTimer, openDoor}
	}
	return state, nil
}
