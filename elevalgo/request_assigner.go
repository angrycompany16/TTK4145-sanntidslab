package elevalgo

import (
	"math"
	"sanntidslab/elevio"
	"slices"
)

const TRAVEL_TIME = 2.0

type ElevatorEntry struct {
	State Elevator
	Id    string
}

// OverView:

// Assignrequest
// Look at the order
// Look at the direction in which everyone is moving
// Obtain sorted list by distance to request floor, iterate
// 1. If moving away from request, skip
//    - To begin with, just take the last elevator if it doesn't work out
// 2. If we reach the end of the list, iterate once more but this time with time
//    to idle

// Seems to work ok
func GetBestElevator(entries []ElevatorEntry, buttonEvent elevio.ButtonEvent) string {
	slices.SortFunc(
		entries,
		func(a, b ElevatorEntry) int {
			return int(math.Abs(float64(a.State.Floor-buttonEvent.Floor)) -
				math.Abs(float64(b.State.Floor-buttonEvent.Floor)))
		},
	)

	for _, entry := range entries {
		if entry.State.Behaviour == idle {
			return entry.Id
		} else if entry.State.direction == -1 && buttonEvent.Floor-entry.State.Floor < 0 {
			return entry.Id
		} else if entry.State.direction == 1 && buttonEvent.Floor-entry.State.Floor > 0 {
			return entry.Id
		}
	}
	// Maybe return the one with the smallest time to idle
	return entries[0].Id
}

// Make a copy of each elevator and use as input, add the new order to each elevator or pass it as a parameter in following func.
// func timeToIdle(e Elevator) float64 {
// 	duration := 0.0

// 	switch e.behaviour {
// 	case idle:
// 		pair := e.chooseDirection()
// 		if pair.dir == stop {
// 			return duration
// 		}
// 	case moving:
// 		duration += TRAVEL_TIME / 2
// 		e.floor += int(e.direction)

// 	case doorOpen:
// 		duration -= e.config.DoorOpenDuration.Seconds() / 2
// 	}

// 	for {
// 		if e.shouldStop() {
// 			e = clearAtCurrentFloor(e)
// 			duration += e.config.DoorOpenDuration.Seconds()

// 			pair := e.chooseDirection()
// 			e.direction = pair.dir
// 			e.behaviour = pair.behaviour
// 			if e.direction == stop {
// 				return duration
// 			}
// 		}
// 		e.floor += int(e.direction)
// 		duration += TRAVEL_TIME
// 	}
// }

// func GetBestOrder(activeElevators []Elevator) []int { //List of functioning elevators. Null aktive heiser exception?
// 	preferredOrderIndices := make([]int, 0, len(activeElevators))
// 	idleTimes := make([]float64, 0, len(activeElevators))

// 	for i, e := range activeElevators {
// 		idleTimes = append(idleTimes, timeToIdle(e))
// 		preferredOrderIndices = append(preferredOrderIndices, i)
// 	}

// 	cmpIdleTimes := func(a, b int) bool { return idleTimes[preferredOrderIndices[a]] < idleTimes[preferredOrderIndices[b]] }
// 	sort.Slice(preferredOrderIndices, cmpIdleTimes)

// 	return preferredOrderIndices
// }
