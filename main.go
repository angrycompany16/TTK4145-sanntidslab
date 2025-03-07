package main

import (
	"flag"
	"fmt"
	distribute "sanntidslab/distribute"
	elevalgo "sanntidslab/elev_al_go"
	timer "sanntidslab/elev_al_go/timer"
	"time"

	"github.com/angrycompany16/driver-go/elevio"
)

const (
	RequestBufferSize   = 1
	defaultElevatorPort = "15657" /* I think? */
)

var (
	localMode bool
)

func main() {
	var port, id string
	flag.StringVar(&port, "port", defaultElevatorPort, "Elevator server port")
	flag.StringVar(&id, "id", "", "Network node id")
	fmt.Println("Started!")

	flag.Parse()

	if port != defaultElevatorPort {
		localMode = true
	}

	elevio.Init("localhost:"+port, elevalgo.NumFloors)
	elevalgo.InitFsm()

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	poll_timer := make(chan bool)
	requests := make(chan distribute.ElevatorRequest, RequestBufferSize)
	networkRequests := make(chan distribute.ElevatorRequest) // Requests coming from other peers

	distribute.InitNode(&elevalgo.ThisElevator, requests, id)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)
	go timer.PollTimer(poll_timer, elevalgo.GetTimeout())
	go distribute.ThisNode.PipeListener(networkRequests)
	go updateLights()

	for {
		select {
		case request := <-networkRequests:
			distribute.ThisNode.TakeRequest(request, true)
		case request := <-requests: // WARNING: sending requests into this channel
			// *will* cause buttons to light up!
			elevalgo.RequestButtonPressed(request.Floor, request.ButtonType)
		case button := <-drv_buttons:
			request := distribute.ElevatorRequest{
				SenderId:   distribute.ThisNode.Id,
				ButtonType: button.Button,
				Floor:      button.Floor,
			}

			distribute.ThisNode.TakeRequest(request, false)
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

// TODO: Maybe put this somewhere else
func updateLights() {
	var sleepTime time.Duration
	if localMode {
		sleepTime = time.Second
	} else {
		sleepTime = time.Millisecond * 10
	}
	for {
		for floor := 0; floor < elevalgo.NumFloors; floor++ {
			for btn := 0; btn < elevalgo.NumButtons; btn++ {
				setLight := elevalgo.GetRequestStatus(floor, btn)
				for _, peer := range distribute.ThisNode.GetPeerList() {
					setLight = setLight || peer.State.Requests[floor][btn]
				}
				elevio.SetButtonLamp(elevio.ButtonType(btn), floor, setLight)
			}
		}
		time.Sleep(sleepTime)
	}
}
