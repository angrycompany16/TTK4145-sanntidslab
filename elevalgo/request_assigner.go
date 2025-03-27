package elevalgo

import (
	"fmt"
	"math"
	"sanntidslab/elevio"
	"slices"
)

type Entry struct {
	State Elevator
	Id    string
}

// Gets the best elevator based on who is closest to the call and who it's most
// convenient for
func GetBestElevator(entries []Entry, buttonEvent elevio.ButtonEvent) string {
	slices.SortFunc(
		entries,
		func(a, b Entry) int {
			return int(math.Abs(float64(a.State.Floor-buttonEvent.Floor)) -
				math.Abs(float64(b.State.Floor-buttonEvent.Floor)))
		},
	)

	for _, entry := range entries {
		fmt.Println("ID:", entry.Id)
		fmt.Println("direction:", entry.State.Direction)
		fmt.Println("state floor:", entry.State.Floor)
		fmt.Println("call floor:", buttonEvent.Floor)
		if (entry.State.Direction == Down && buttonEvent.Floor-entry.State.Floor+1 > 0) ||
			(entry.State.Direction == Up && buttonEvent.Floor-entry.State.Floor-1 < 0) {
			continue
		}
		return entry.Id
	}

	return entries[0].Id
}
