package networking

import (
	"math"
	"sanntidslab/elevalgo"
	"sanntidslab/elevio"
	"slices"
)

type ElevatorEntry struct {
	State elevalgo.Elevator
	Id    string
}

// Gets the best elevator based on who is closest to the call and who it's most
// convenient for
func GetBestElevator(entries []ElevatorEntry, buttonEvent elevio.ButtonEvent) string {
	slices.SortFunc(
		entries,
		func(a, b ElevatorEntry) int {
			return int(math.Abs(float64(a.State.Floor-buttonEvent.Floor)) -
				math.Abs(float64(b.State.Floor-buttonEvent.Floor)))
		},
	)

	for _, entry := range entries {
		if (entry.State.Direction == elevalgo.Down && buttonEvent.Floor-entry.State.Floor+1 > 0) ||
			(entry.State.Direction == elevalgo.Up && buttonEvent.Floor-entry.State.Floor-1 < 0) {
			continue
		}
		return entry.Id
	}

	return entries[0].Id
}
