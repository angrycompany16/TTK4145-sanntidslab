package elevalgo

import (
	"sanntidslab/elevio"
)

func (e *Elevator) requestsAbove() bool {
	for f := e.Floor + 1; f < NumFloors; f++ {
		for btn := range NumButtons {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func (e *Elevator) requestsBelow() bool {
	for f := range e.Floor {
		for btn := range NumButtons {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func (e *Elevator) requestsHere() bool {
	for btn := range NumButtons {
		if e.Requests[e.Floor][btn] {
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
	case stop:
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
		return e.Requests[e.Floor][elevio.BT_HallDown] || e.Requests[e.Floor][elevio.BT_Cab] || !e.requestsBelow()
	case up:
		return e.Requests[e.Floor][elevio.BT_HallUp] || e.Requests[e.Floor][elevio.BT_Cab] || !e.requestsAbove()
	default:
		return true
	}
}

func (e *Elevator) shouldClearImmediately(buttonFloor int, buttonType elevio.ButtonType) bool {
	switch e.config.ClearRequestVariant {
	case clearAll:
		return e.Floor == buttonFloor
	case clearSameDir:
		return e.Floor == buttonFloor && ((e.direction == up && buttonType == elevio.BT_HallUp) ||
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
		for btn := range NumButtons {
			e.Requests[e.Floor][btn] = false
		}

	case clearSameDir:
		e.Requests[e.Floor][elevio.BT_Cab] = false
		switch e.direction {
		case up:
			if !e.requestsAbove() && !e.Requests[e.Floor][elevio.BT_HallUp] {
				e.Requests[e.Floor][elevio.BT_HallDown] = false
			}
			e.Requests[e.Floor][elevio.BT_HallUp] = false
		case down:
			if !e.requestsBelow() && !e.Requests[e.Floor][elevio.BT_HallDown] {
				e.Requests[e.Floor][elevio.BT_HallUp] = false
			}
			e.Requests[e.Floor][elevio.BT_HallDown] = false
		default:
			e.Requests[e.Floor][elevio.BT_HallUp] = false
			e.Requests[e.Floor][elevio.BT_HallDown] = false
		}
	}

	return e
}
