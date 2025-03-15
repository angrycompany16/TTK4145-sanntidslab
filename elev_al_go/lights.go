package elevalgo

import (
	"github.com/angrycompany16/driver-go/elevio"
)

var (
	cachedLightsState = LightsState{}
)

type LightsState struct {
	Lights [NumFloors][NumButtons]bool
}

func (e *Elevator) SetLights(lightsState LightsState) {
	// No point in setting the lights to what they already are
	if lightsState == cachedLightsState {
		return
	}
	cachedLightsState = lightsState

	for floor := 0; floor < NumFloors; floor++ {
		for btn := 0; btn < NumButtons; btn++ {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, lightsState.Lights[floor][btn])
		}
	}
}

func MergeHallLights(localElevator Elevator, allElevators []Elevator) (lightsState LightsState) {
	for floor := 0; floor < NumFloors; floor++ {
		// Note: This only works because hall buttons come first
		// Sets hall buttons based on global elevator states,
		// Set cab buttons based on local elevator states.
		for btn := 0; btn < numHallButtons; btn++ {
			for _, elevator := range allElevators {
				lightsState.Lights[floor][btn] = lightsState.Lights[floor][btn] || elevator.Requests[floor][btn]
			}
		}
		for btn := numHallButtons; btn < NumButtons; btn++ {
			lightsState.Lights[floor][btn] = localElevator.Requests[floor][btn]
		}
	}
	return
}
