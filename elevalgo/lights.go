package elevalgo

// import (
// 	"sanntidslab/elevalgo"

// 	"github.com/angrycompany16/driver-go/elevio"
// )

// var (
// 	cachedLightsState = LightsState{}
// )

// type LightsState struct {
// 	Lights [NumFloors][NumButtons]bool
// }

// func setCabLights() {

// }

// func setHallLights(state elevalgo.Elevator, peerStates []Elevator) {
// 	for i := range NumFloors {
// 		for j := range numHallButtons {
// 			for _, peerState := range peerStates {
// 				LightsState.Lights[i][j] = LightsState.Lights[i][j] || peerState.Requests[i][j]

// 			}
// 		}
// 	}
// }

// func SetLights(lightsState LightsState) {
// 	// TODO: Should not be needed?
// 	// No point in setting the lights to what they already are
// 	if lightsState == cachedLightsState {
// 		return
// 	}
// 	cachedLightsState = lightsState

// 	for floor := range NumFloors {
// 		for btn := range NumButtons {
// 			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, lightsState.Lights[floor][btn])
// 		}
// 	}
// }

// func MergeHallLights(localElevator Elevator, allElevators []Elevator) (lightsState LightsState) {
// 	for floor := range NumFloors {
// 		// Sets hall buttons based on global elevator states,
// 		// sets cab buttons based on local elevator states.
// 		for btn := range numHallButtons {
// 			for _, elevator := range allElevators {
// 				lightsState.Lights[floor][btn] = lightsState.Lights[floor][btn] || elevator.Requests[floor][btn]
// 			}
// 		}
// 		for btn := numHallButtons; btn < NumButtons; btn++ {
// 			lightsState.Lights[floor][btn] = localElevator.Requests[floor][btn]
// 		}
// 	}
// 	return
// }
