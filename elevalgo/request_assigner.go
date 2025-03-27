package elevalgo

import (
	"math"
	"sanntidslab/elevio"
	"slices"
)

type ElevatorEntry struct {
	State Elevator
	Id    string
}

// Sort all elevators by distance to request floor
// If idle, assign
// If moving towards request, assign
// If all elevators are moving away from request, assign closest

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

	return entries[0].Id
}
