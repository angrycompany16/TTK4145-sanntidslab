package elevalgo

func (e *Elevator) RequestsAbove() bool {
	for f := e.floor + 1; f < NUM_FLOORS; f++ {
		for btn := 0; btn < NUM_BUTTONS; btn++ {
			if e.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func (e *Elevator) RequestsBelow() bool {
	for f := 0; f < e.floor; f++ {
		for btn := 0; btn < NUM_BUTTONS; btn++ {
			if e.requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func (e *Elevator) RequestsHere() bool {
	for btn := 0; btn < NUM_BUTTONS; btn++ {
		if e.requests[e.floor][btn] {
			return true
		}
	}
	return false
}

func (e *Elevator) RequestsChooseDirection() DirBehaviourPair {
	switch e.direction {
	case DIR_UP:
		if e.RequestsAbove() {
			return DirBehaviourPair{DIR_UP, BEHAVIOUR_MOVING}
		} else if e.RequestsHere() {
			return DirBehaviourPair{DIR_STOP, BEHAVIOUR_DOOR_OPEN}
		} else if e.RequestsBelow() {
			return DirBehaviourPair{DIR_DOWN, BEHAVIOUR_MOVING}
		} else {
			return DirBehaviourPair{DIR_STOP, BEHAVIOUR_IDLE}
		}
	case DIR_DOWN:
		if e.RequestsBelow() {
			return DirBehaviourPair{DIR_DOWN, BEHAVIOUR_MOVING}
		} else if e.RequestsHere() {
			return DirBehaviourPair{DIR_STOP, BEHAVIOUR_DOOR_OPEN}
		} else if e.RequestsAbove() {
			return DirBehaviourPair{DIR_UP, BEHAVIOUR_MOVING}
		} else {
			return DirBehaviourPair{DIR_STOP, BEHAVIOUR_IDLE}
		}
	case DIR_STOP: // there should only be one request in the Stop case. Checking up or down first is arbitrary.
		if e.RequestsHere() {
			return DirBehaviourPair{DIR_STOP, BEHAVIOUR_DOOR_OPEN}
		} else if e.RequestsAbove() {
			return DirBehaviourPair{DIR_UP, BEHAVIOUR_MOVING}
		} else if e.RequestsBelow() {
			return DirBehaviourPair{DIR_DOWN, BEHAVIOUR_MOVING}
		} else {
			return DirBehaviourPair{DIR_STOP, BEHAVIOUR_IDLE}
		}
	default:
		return DirBehaviourPair{DIR_STOP, BEHAVIOUR_IDLE}
	}
}

func (e *Elevator) RequestsShouldStop() bool {
	switch e.direction {
	case DIR_DOWN:
		return e.requests[e.floor][BTN_HALLDOWN] || e.requests[e.floor][BTN_HALLCAB] || !e.RequestsBelow()
	case DIR_UP:
		return e.requests[e.floor][BTN_HALLUP] || e.requests[e.floor][BTN_HALLCAB] || !e.RequestsAbove()
	default:
		return true
	}
}

func (e *Elevator) RequestsShouldClearImmediately(buttonFloor int, buttonType Button) bool {
	switch e.config.clearRequestVariation {
	case CV_All:
		return e.floor == buttonFloor
	case CV_InDirn:
		return e.floor == buttonFloor && ((e.direction == DIR_UP && buttonType == BTN_HALLUP) ||
			(e.direction == DIR_DOWN && buttonType == BTN_HALLDOWN) ||
			e.direction == DIR_STOP ||
			buttonType == BTN_HALLCAB)
	default:
		return false
	}
}

func RequestsClearAtCurrentFloor(e Elevator) Elevator {
	switch e.config.clearRequestVariation {
	case CV_All:
		for btn := 0; btn < NUM_BUTTONS; btn++ {
			e.requests[e.floor][btn] = false
		}

	case CV_InDirn:
		e.requests[e.floor][BTN_HALLCAB] = false
		switch e.direction {
		case DIR_UP:
			if !e.RequestsAbove() && !e.requests[e.floor][BTN_HALLUP] {
				e.requests[e.floor][BTN_HALLDOWN] = false
			}
			e.requests[e.floor][BTN_HALLUP] = false
		case DIR_DOWN:
			if !e.RequestsBelow() && !e.requests[e.floor][BTN_HALLDOWN] {
				e.requests[e.floor][BTN_HALLUP] = false
			}
			e.requests[e.floor][BTN_HALLDOWN] = false
		default:
			e.requests[e.floor][BTN_HALLUP] = false
			e.requests[e.floor][BTN_HALLDOWN] = false
		}
	}

	return e
}
