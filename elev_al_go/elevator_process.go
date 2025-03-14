package elevalgo

import (
	requests "sanntidslab/p2p/requests"
	timer "sanntidslab/elev_al_go/timer"
)

func ElevatorProcess(
	/* input only channels */
	floorChan <-chan int,
	obstructionChan <-chan bool,
	orderChan <-chan requests.RequestInfo,
	/* output only channels */
	elevatorStateChan chan<- Elevator,
) {

	for {
		select {
		case requestInfo := <-orderChan:
			requestButtonPressed(requestInfo.Floor, requestInfo.ButtonType)
		case floor := <-floorChan:
			onFloorArrival(floor)
		case obstructionEvent := <-obstructionChan:
			doorObstructed(obstructionEvent)
		case <-timer.TimeoutChan:
	 		timer.StopTimer()
	 		onDoorTimeout()
		default:
			elevatorStateChan <- GetState()
		}
	}
}