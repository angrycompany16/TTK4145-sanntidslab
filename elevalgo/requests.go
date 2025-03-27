package elevalgo

import (
	"sanntidslab/elevio"
)

func hasRequestsAbove(elevator Elevator) bool {
	for f := elevator.Floor + 1; f < NumFloors; f++ {
		for btn := range NumButtons {
			if elevator.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func hasRequestsBelow(elevator Elevator) bool {
	for f := range elevator.Floor {
		for btn := range NumButtons {
			if elevator.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func hasRequestsHere(elevator Elevator) bool {
	for btn := range NumButtons {
		if elevator.Requests[elevator.Floor][btn] {
			return true
		}
	}
	return false
}

func chooseDirection(elevator Elevator) dirBehaviourPair {
	switch elevator.Direction {
	case Up:
		if hasRequestsAbove(elevator) {
			return dirBehaviourPair{Up, moving}
		} else if hasRequestsHere(elevator) {
			return dirBehaviourPair{Stop, doorOpen}
		} else if hasRequestsBelow(elevator) {
			return dirBehaviourPair{Down, moving}
		} else {
			return dirBehaviourPair{Stop, idle}
		}
	case Down:
		if hasRequestsBelow(elevator) {
			return dirBehaviourPair{Down, moving}
		} else if hasRequestsHere(elevator) {
			return dirBehaviourPair{Stop, doorOpen}
		} else if hasRequestsAbove(elevator) {
			return dirBehaviourPair{Up, moving}
		} else {
			return dirBehaviourPair{Stop, idle}
		}
	case Stop: // Note: there should only be one request in the Stop case. Checking up or down first is arbitrary.
		if hasRequestsHere(elevator) {
			return dirBehaviourPair{Stop, doorOpen}
		} else if hasRequestsAbove(elevator) {
			return dirBehaviourPair{Up, moving}
		} else if hasRequestsBelow(elevator) {
			return dirBehaviourPair{Down, moving}
		} else {
			return dirBehaviourPair{Stop, idle}
		}
	default:
		return dirBehaviourPair{Stop, idle}
	}
}

func shouldStop(elevator Elevator) bool {
	switch elevator.Direction {
	case Down:
		return elevator.Requests[elevator.Floor][elevio.BT_HallDown] || elevator.Requests[elevator.Floor][elevio.BT_Cab] || !hasRequestsBelow(elevator)
	case Up:
		return elevator.Requests[elevator.Floor][elevio.BT_HallUp] || elevator.Requests[elevator.Floor][elevio.BT_Cab] || !hasRequestsAbove(elevator)
	default:
		return true
	}
}

func shouldClearImmediately(elevator Elevator, buttonFloor int, buttonType elevio.ButtonType) bool {
	switch elevator.config.ClearRequestVariant {
	case clearAll:
		return elevator.Floor == buttonFloor
	case clearSameDir:
		return elevator.Floor == buttonFloor && ((elevator.Direction == Up && buttonType == elevio.BT_HallUp) ||
			(elevator.Direction == Down && buttonType == elevio.BT_HallDown) ||
			elevator.Direction == Stop ||
			buttonType == elevio.BT_Cab)
	default:
		return false
	}
}

func clearAtCurrentFloor(elevator Elevator) Elevator {
	switch elevator.config.ClearRequestVariant {
	case clearAll:
		for btn := range NumButtons {
			elevator.Requests[elevator.Floor][btn] = false
		}

	case clearSameDir:
		elevator.Requests[elevator.Floor][elevio.BT_Cab] = false
		switch elevator.Direction {
		case Up:
			if !hasRequestsAbove(elevator) && !elevator.Requests[elevator.Floor][elevio.BT_HallUp] {
				elevator.Requests[elevator.Floor][elevio.BT_HallDown] = false
			}
			elevator.Requests[elevator.Floor][elevio.BT_HallUp] = false
		case Down:
			if !hasRequestsBelow(elevator) && !elevator.Requests[elevator.Floor][elevio.BT_HallDown] {
				elevator.Requests[elevator.Floor][elevio.BT_HallUp] = false
			}
			elevator.Requests[elevator.Floor][elevio.BT_HallDown] = false
		default:
			elevator.Requests[elevator.Floor][elevio.BT_HallUp] = false
			elevator.Requests[elevator.Floor][elevio.BT_HallDown] = false
		}
	}

	return elevator
}
