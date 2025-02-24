package main

import (
	"flag"
	"fmt"
	elevalgo "sanntidslab/elev_al_go"
	timer "sanntidslab/elev_al_go/timer"
	networking "sanntidslab/network"

	"github.com/angrycompany16/driver-go/elevio"
)

const (
	normalMode  = "normal"
	virtualMode = "virtual"
)

func main() {
	var mode string
	flag.StringVar(&mode, "mode", normalMode, "The mode in which to run the elevator")
	// fmt.Println("Started!")

	flag.Parse()

	if mode == virtualMode {
		elevator := elevalgo.MakeUninitializedelevator()
		networking.InitElevator(&elevator)
		for {
			if networking.ThisNode.GetDebugInput() {
				fmt.Println("Exiting")
				return
			}
		}
	}

	elevio.Init("localhost:15657", elevalgo.NumFloors)
	elevalgo.InitFsm()
	networking.InitElevator(&elevalgo.ThisElevator)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	poll_timer := make(chan bool)
	incoming_requests := make(chan networking.ElevatorRequest)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go timer.PollTimer(poll_timer, elevalgo.GetTimeout())
	go networking.ThisNode.PipeListener(incoming_requests)

	for {
		select {
		case request := <-incoming_requests:
			elevalgo.RequestButtonPressed(request.Floor, request.ButtonType)
		case button := <-drv_buttons:
			// Find peer which should take the request
			// Send the request
			// Note that non physical elevators cannot send messages, only receive
			networking.ThisNode.SendMsg(button.Button, button.Floor)
			// elevalgo.RequestButtonPressed(button.Floor, button.Button)
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
