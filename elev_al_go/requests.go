package elevalgo

func (e *elevator) requestsAbove() bool {
	for f := e.floor + 1; f < NumFloors; f++ {
		for btn := 0; btn < NumButtons; btn++ {
			if e.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func (e *elevator) requestsBelow() bool {
	for f := 0; f < e.floor; f++ {
		for btn := 0; btn < NumButtons; btn++ {
			if e.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func (e *elevator) requestsHere() bool {
	for btn := 0; btn < NumButtons; btn++ {
		if e.requests[e.floor][btn] {
			return true
		}
	}
	return false
}

func (e *elevator) chooseDirection() dirBehaviourPair {
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

func (e *elevator) shouldStop() bool {
	switch e.direction {
	case down:
		return e.requests[e.floor][hallDown] || e.requests[e.floor][hallCab] || !e.requestsBelow()
	case up:
		return e.requests[e.floor][hallUp] || e.requests[e.floor][hallCab] || !e.requestsAbove()
	default:
		return true
	}
}

func (e *elevator) shouldClearImmediately(buttonFloor int, buttonType Button) bool {
	switch e.config.ClearRequestVariant {
	case clearAll:
		return e.floor == buttonFloor
	case clearSameDir:
		return e.floor == buttonFloor && ((e.direction == up && buttonType == hallUp) ||
			(e.direction == down && buttonType == hallDown) ||
			e.direction == stop ||
			buttonType == hallCab)
	default:
		return false
	}
}

func clearAtCurrentFloor(e elevator) elevator {
	switch e.config.ClearRequestVariant {
	case clearAll:
		for btn := 0; btn < NumButtons; btn++ {
			e.requests[e.floor][btn] = false
		}

	case clearSameDir:
		e.requests[e.floor][hallCab] = false
		switch e.direction {
		case up:
			if !e.requestsAbove() && !e.requests[e.floor][hallUp] {
				e.requests[e.floor][hallDown] = false
			}
			e.requests[e.floor][hallUp] = false
		case down:
			if !e.requestsBelow() && !e.requests[e.floor][hallDown] {
				e.requests[e.floor][hallUp] = false
			}
			e.requests[e.floor][hallDown] = false
		default:
			e.requests[e.floor][hallUp] = false
			e.requests[e.floor][hallDown] = false
		}
	}

	return e
}
