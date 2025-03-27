package door

import (
	"fmt"
	"sanntidslab/elevio"
)

type state int

const (
	open state = iota
	closed
	stuck
)

type command int

const (
	openDoor command = iota
	closeDoor
	resetTimer
	resetObstructionTimer
	stopObstructionTimer
)

// Runs a simple state machine for the elevator door that accepts requests and handles
// obstruction events. It also manages a timer that panics if the door remains
// obstructed. Note that the obstruction panic timer will not start until the door opens.
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
	var doorInstanceState state
	initCommands := []command{closeDoor}
	if startObstructed {
		doorInstanceState = stuck
		initCommands = append(initCommands, resetObstructionTimer)
	} else {
		doorInstanceState = closed
		initCommands = append(initCommands, stopObstructionTimer)
	}
	executeCommands(startDoorTimerChan, doorCloseChan, resetObstructionTimerChan, stopObstructionTimerChan, initCommands)

	for {
		var commands []command
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
	commands []command,
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

func onObstructionEvent(obstructionEvent bool, state state) (newState state, commands []command) {
	if obstructionEvent && state != closed {
		commands = append(commands, resetObstructionTimer)
	} else {
		commands = append(commands, stopObstructionTimer)
	}

	switch state {
	case open:
		newState = stuck
	case closed:
		newState = stuck
	case stuck:
		if !obstructionEvent {
			fmt.Println("Obstruction freed")
			newState = open
			commands = append(commands, resetTimer)
			return
		}
		newState = stuck
	}
	return
}

func onDoorTimeout(state state) (newState state, commands []command) {
	switch state {
	case open:
		newState = closed
		commands = append(commands, closeDoor)
	case closed:
	case stuck:
		commands = append(commands, resetTimer)
	}
	return
}

func onDoorRequest(state state) (newState state, commands []command) {
	switch state {
	case open:
		commands = append(commands, resetTimer)
	case closed:
		newState = open
		commands = append(commands, resetTimer, openDoor)
	case stuck:
		newState = stuck
		commands = append(commands, openDoor, resetObstructionTimer)
	}
	return
}
