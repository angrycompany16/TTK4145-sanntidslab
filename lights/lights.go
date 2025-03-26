package lights

import (
	"sanntidslab/elevalgo"
	"sanntidslab/elevio"
)

type LightsState struct {
	Lights [elevalgo.NumFloors][elevalgo.NumButtons]bool
}

// Problem:
// Lights aren't disabled when redistributed backup calls are taken
// This likely happens because

// Sets lights based on inputs from elevator and peers
func RunLights(
	elevatorStateChan <-chan elevalgo.Elevator,
	peerListChan <-chan []elevalgo.Elevator,
) {
	lightsState := LightsState{}
	peerList := make([]elevalgo.Elevator, 0)
	state := elevalgo.Elevator{}

	// Clear lights
	setLights(lightsState)

	for {
		var newLightState LightsState
		select {
		case newElevator := <-elevatorStateChan:
			state = newElevator
			newLightState = getLights(state, append(peerList, state))
		case newPeerList := <-peerListChan:
			peerList = newPeerList
			newLightState = getLights(state, append(peerList, state))
		}

		if newLightState != lightsState {
			setLights(newLightState)
			lightsState = newLightState
		}
	}
}

func setLights(lightsState LightsState) {
	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			elevio.SetButtonLamp(elevio.ButtonType(j), i, lightsState.Lights[i][j])
		}
	}
}

func getLights(state elevalgo.Elevator, allStates []elevalgo.Elevator) (newLightsState LightsState) {
	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumHallButtons {
			for _, peerState := range allStates {
				newLightsState.Lights[i][j] = newLightsState.Lights[i][j] || peerState.Requests[i][j]
			}
		}

		for j := elevalgo.NumHallButtons; j < elevalgo.NumButtons; j++ {
			newLightsState.Lights[i][j] = state.Requests[i][j]
		}
	}
	return
}
