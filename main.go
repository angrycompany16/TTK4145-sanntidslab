package main

import (
	"flag"
	"fmt"
	distribute "sanntidslab/distribute"
	elevalgo "sanntidslab/elev_al_go"
	timer "sanntidslab/elev_al_go/timer"

	"github.com/angrycompany16/driver-go/elevio"
)

const (
	RequestBufferSize = 1
)

func main() {
	var port, id string
	flag.StringVar(&port, "port", "", "Elevator server port")
	flag.StringVar(&id, "id", "", "Network node id")
	fmt.Println("Started!")

	flag.Parse()

	elevio.Init("localhost:"+port, elevalgo.NumFloors)
	elevalgo.InitFsm()

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	poll_timer := make(chan bool)
	requests := make(chan distribute.ElevatorRequest, RequestBufferSize) // Important: This needs to
	acks := make(chan distribute.Ack, 1)
	records := make(chan distribute.Record, 1)

	distribute.InitNode(&elevalgo.ThisElevator, requests, id)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go timer.PollTimer(poll_timer, elevalgo.GetTimeout())
	go distribute.ThisNode.PipeListener(requests, acks, records)
	// go distribute.ThisNode.LocalRequests(requests)

	// New backup idea: Every node already knows the last broadcasted state of every other node
	// Upon connecting, read the state we are supposed to have from this node
	// If there are calls we haven't taken, service these calls
	// This should make things easier

	// Problem: The backup needs to know when a call has been serviced, otherwise it's going
	// to screw up and deliver requests multiple times
	// This is getting scarily close to being a pure p2p solution
	for {
		select {
		case request := <-requests: // WARNING: sending requests into this channel
			// *will* cause buttons to light up!
			elevalgo.RequestButtonPressed(request.Floor, request.ButtonType)
		case button := <-drv_buttons:
			// Store the call and broadcast it
			// As soon as someone broadcasts a worldview that shows they have the
			// same requests, consider this as a guarantee and take the request

			distribute.ThisNode.TakeRequest(button)

			// if distribute.ThisNode.SendRequest(button) {
			// 	requests <- distribute.ThisNode.SelfRequestNode(button)
			// 	fmt.Println("Requested from self")
			// }
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
