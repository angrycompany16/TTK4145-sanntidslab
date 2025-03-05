package elevalgo

import "github.com/angrycompany16/driver-go/elevio"

func (e *Elevator) requestsAbove() bool {
	for f := e.floor + 1; f < NumFloors; f++ {
		for btn := 0; btn < NumButtons; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func (e *Elevator) requestsBelow() bool {
	for f := 0; f < e.floor; f++ {
		for btn := 0; btn < NumButtons; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func (e *Elevator) requestsHere() bool {
	for btn := 0; btn < NumButtons; btn++ {
		if e.Requests[e.floor][btn] {
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
		return e.Requests[e.floor][elevio.BT_HallDown] || e.Requests[e.floor][elevio.BT_Cab] || !e.requestsBelow()
	case up:
		return e.Requests[e.floor][elevio.BT_HallUp] || e.Requests[e.floor][elevio.BT_Cab] || !e.requestsAbove()
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
			e.Requests[e.floor][btn] = false
		}

	case clearSameDir:
		e.Requests[e.floor][elevio.BT_Cab] = false
		switch e.direction {
		case up:
			if !e.requestsAbove() && !e.Requests[e.floor][elevio.BT_HallUp] {
				e.Requests[e.floor][elevio.BT_HallDown] = false
			}
			e.Requests[e.floor][elevio.BT_HallUp] = false
		case down:
			if !e.requestsBelow() && !e.Requests[e.floor][elevio.BT_HallDown] {
				e.Requests[e.floor][elevio.BT_HallUp] = false
			}
			e.Requests[e.floor][elevio.BT_HallDown] = false
		default:
			e.Requests[e.floor][elevio.BT_HallUp] = false
			e.Requests[e.floor][elevio.BT_HallDown] = false
		}
	}

	return e
}
