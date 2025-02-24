package elevalgo

import "github.com/angrycompany16/driver-go/elevio"

func (e *Elevator) requestsAbove() bool {
	for f := e.floor + 1; f < NumFloors; f++ {
		for btn := 0; btn < NumButtons; btn++ {
			if e.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func (e *Elevator) requestsBelow() bool {
	for f := 0; f < e.floor; f++ {
		for btn := 0; btn < NumButtons; btn++ {
			if e.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func (e *Elevator) requestsHere() bool {
	for btn := 0; btn < NumButtons; btn++ {
		if e.requests[e.floor][btn] {
			return true
		}
	}
	return false
}

func (e *Elevator) chooseDirection() dirBehaviourPair {
	switch e.direction {
	case up:
		if e.requestsAbove() {
			return dirBehaviourPair{up, moving}
		} else if e.requestsHere() {
			return dirBehaviourPair{stop, doorOpen}
		} else if e.requestsBelow() {
			return dirBehaviourPair{down, moving}
		} else {
			return dirBehaviourPair{stop, idle}
		}
	case down:
		if e.requestsBelow() {
			return dirBehaviourPair{down, moving}
		} else if e.requestsHere() {
			return dirBehaviourPair{stop, doorOpen}
		} else if e.requestsAbove() {
			return dirBehaviourPair{up, moving}
		} else {
			return dirBehaviourPair{stop, idle}
		}
	case stop: // there should only be one request in the Stop case. Checking up or down first is arbitrary.
		if e.requestsHere() {
			return dirBehaviourPair{stop, doorOpen}
		} else if e.requestsAbove() {
			return dirBehaviourPair{up, moving}
		} else if e.requestsBelow() {
			return dirBehaviourPair{down, moving}
		} else {
			return dirBehaviourPair{stop, idle}
		}
	default:
		return dirBehaviourPair{stop, idle}
	}
}

func (e *Elevator) shouldStop() bool {
	switch e.direction {
	case down:
		return e.requests[e.floor][elevio.BT_HallDown] || e.requests[e.floor][elevio.BT_Cab] || !e.requestsBelow()
	case up:
		return e.requests[e.floor][elevio.BT_HallUp] || e.requests[e.floor][elevio.BT_Cab] || !e.requestsAbove()
	default:
		return true
	}
}

func (e *Elevator) shouldClearImmediately(buttonFloor int, buttonType elevio.ButtonType) bool {
	switch e.config.ClearRequestVariant {
	case clearAll:
		return e.floor == buttonFloor
	case clearSameDir:
		return e.floor == buttonFloor && ((e.direction == up && buttonType == elevio.BT_HallUp) ||
			(e.direction == down && buttonType == elevio.BT_HallDown) ||
			e.direction == stop ||
			buttonType == elevio.BT_Cab)
	default:
		return false
	}
}

func clearAtCurrentFloor(e Elevator) Elevator {
	switch e.config.ClearRequestVariant {
	case clearAll:
		for btn := 0; btn < NumButtons; btn++ {
			e.requests[e.floor][btn] = false
		}

	case clearSameDir:
		e.requests[e.floor][elevio.BT_Cab] = false
		switch e.direction {
		case up:
			if !e.requestsAbove() && !e.requests[e.floor][elevio.BT_HallUp] {
				e.requests[e.floor][elevio.BT_HallDown] = false
			}
			e.requests[e.floor][elevio.BT_HallUp] = false
		case down:
			if !e.requestsBelow() && !e.requests[e.floor][elevio.BT_HallDown] {
				e.requests[e.floor][elevio.BT_HallUp] = false
			}
			e.requests[e.floor][elevio.BT_HallDown] = false
		default:
			e.requests[e.floor][elevio.BT_HallUp] = false
			e.requests[e.floor][elevio.BT_HallDown] = false
		}
	}

	return e
}
