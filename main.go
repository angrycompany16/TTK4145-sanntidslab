package main

import (
	"fmt"
	elevalgo "sanntidslab/elev_al_go"
	timer "sanntidslab/elev_al_go/timer"

	"github.com/angrycompany16/driver-go/elevio"
)

func main() {
	fmt.Println("Started!")

	elevio.Init("localhost:15657", elevalgo.NumFloors)

	elevalgo.InitFsm()

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	poll_timer := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go timer.PollTimer(poll_timer, elevalgo.GetTimeout())

	for {
		select {
		case button := <-drv_buttons:
			elevalgo.RequestButtonPressed(button.Floor, elevalgo.Button(button.Button))
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
