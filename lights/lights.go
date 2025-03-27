package lights

import (
	"sanntidslab/elevalgo"
	"sanntidslab/elevio"
)

type lightsState struct {
	lights [elevalgo.NumFloors][elevalgo.NumButtons]bool
}

// Sets lights based on inputs from elevator and peers
func RunLights(
	elevatorStateChan <-chan elevalgo.Elevator,
	peerListChan <-chan []elevalgo.Elevator,
) {
	lightsState := lightsState{}
	peerList := make([]elevalgo.Elevator, 0)
	elevator := elevalgo.Elevator{}

	setLights(lightsState)

	for {
		select {
		case elevator = <-elevatorStateChan:
			lightsState = getLights(elevator, append(peerList, elevator))
		case peerList = <-peerListChan:
			lightsState = getLights(elevator, append(peerList, elevator))
		}

		setLights(lightsState)
	}
}

func setLights(lightsState lightsState) {
	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			elevio.SetButtonLamp(elevio.ButtonType(j), i, lightsState.lights[i][j])
		}
	}
}

func getLights(state elevalgo.Elevator, allStates []elevalgo.Elevator) (newLightsState lightsState) {
	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumHallButtons {
			for _, peerState := range allStates {
				newLightsState.lights[i][j] = newLightsState.lights[i][j] || peerState.Requests[i][j]
			}
		}

		for j := elevalgo.NumHallButtons; j < elevalgo.NumButtons; j++ {
			newLightsState.lights[i][j] = state.Requests[i][j]
		}
	}
	return
}
