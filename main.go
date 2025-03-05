package main

import (
	"sanntidslab/backup"
	elevalgo "sanntidslab/elev_al_go"
	timer "sanntidslab/elev_al_go/timer"

	// networking "sanntidslab/network"

	"github.com/angrycompany16/driver-go/elevio"
)

func main() {
	elevio.Init("localhost:15657", elevalgo.NumFloors)
	elevalgo.InitFsm()
	backup.InitElevator(&elevalgo.ThisElevator)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	poll_timer := make(chan bool)
	incoming_requests := make(chan backup.ElevatorRequest)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go timer.PollTimer(poll_timer, elevalgo.GetTimeout())
	go backup.ThisNode.PipeListener(incoming_requests)

	for {
		select {
		case request := <-incoming_requests:
			elevalgo.RequestButtonPressed(request.Floor, request.ButtonType)
		case button := <-drv_buttons:
			// Find peer which should take the request
			// Send the request
			backup.ThisNode.SendMsg(button.Button, button.Floor)
		case floor := <-drv_floors:
			elevalgo.OnFloorArrival(floor)
		case obstructed := <-drv_obstr:
			if obstructed {
				elevalgo.DoorObstructed()
			}
		case <-poll_timer:
			timer.StopTimer()
			elevalgo.OnDoorTimeout()
		}
	}
}
