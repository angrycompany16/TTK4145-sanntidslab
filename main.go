package main

import (
	"fmt"
	elevalgo "sanntidslab/elev_al_go"

	"github.com/angrycompany16/driver-go/elevio"
)

func main() {
	fmt.Println("Started!")

	elevio.Init("localhost:15657", elevalgo.NUM_FLOORS)

	elevalgo.MakeFsm()

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	poll_timer := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go elevalgo.PollTimer(poll_timer)

	for {
		select {
		case button := <-drv_buttons:
			elevalgo.FsmOnRequestButtonPress(button.Floor, elevalgo.Button(button.Button))
		case floor := <-drv_floors:
			elevalgo.FsmOnFloorArrival(floor)
		case obstructed := <-drv_obstr:
			if obstructed {
				elevalgo.DoorObstructed()
			}
		case <-poll_timer:
			elevalgo.StopTimer()
			elevalgo.FsmOnDoorTimeout()
		}
	}
}
