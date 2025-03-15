package elevalgo

import (
	"sort"
)

const TRAVEL_TIME = 1.0

// TODO: it may be possible to do this better. One could compute how far each elevator is
// from being able to reach the request, and order based on that

// Make a copy of each elevator and use as input, add the new order to each elevator or pass it as a parameter in following func.
func timeToIdle(e Elevator) float64 {
	duration := 0.0
	// doorOpenTime := 3 //int(e.config.DoorOpenDuration)

	switch e.behaviour {
	case idle:
		pair := e.chooseDirection()
		if pair.dir == stop {
			return duration
		}
	case moving:
		duration += TRAVEL_TIME / 2
		e.floor += int(e.direction)

	case doorOpen:
		duration -= e.config.DoorOpenDuration.Seconds() / 2
	}

	for {
		if e.shouldStop() {
			e = clearAtCurrentFloor(e)
			duration += e.config.DoorOpenDuration.Seconds()

			pair := e.chooseDirection()
			e.direction = pair.dir
			e.behaviour = pair.behaviour
			if e.direction == stop {
				return duration
			}
		}
		e.floor += int(e.direction)
		duration += TRAVEL_TIME
	}
}

func GetBestOrder(activeElevators []Elevator) []int { //List of functioning elevators. Null aktive heiser exception?
	preferredOrderIndices := make([]int, 0, len(activeElevators))
	idleTimes := make([]float64, 0, len(activeElevators))

	for i, e := range activeElevators {
		idleTimes = append(idleTimes, timeToIdle(e))
		preferredOrderIndices = append(preferredOrderIndices, i)
	}

	cmpIdleTimes := func(a, b int) bool { return idleTimes[preferredOrderIndices[a]] < idleTimes[preferredOrderIndices[b]] }
	sort.Slice(preferredOrderIndices, cmpIdleTimes)

	return preferredOrderIndices
}
