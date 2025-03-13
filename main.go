package main

import (
	"flag"
	"fmt"
	elevalgo "sanntidslab/elev_al_go"
	"sanntidslab/elev_al_go/timer"
	"sanntidslab/p2p"
	"sanntidslab/p2p/requests"
	"strconv"

	"github.com/angrycompany16/Network-go/network/broadcast"
	"github.com/angrycompany16/driver-go/elevio"
)

const (
	RequestBufferSize   = 1
	defaultElevatorPort = 15657 /* I think? */
	stateBroadcastPort  = 36251 // Akkordrekke
)

// TODO: *Read* code complete checklist properly and at least try to make the code
// quality good

// TODO: It seems that network.go sometimes crashes on startup with an unbelievably
// long stack trace... that's probably not a great thing
// Found out that it's due to simultaneous read and write from map

// TODO: Implement the backup actually taking lost requests itself
// TODO: In case of disconnect, all requests should also be taken
// TODO: Arbitration/priority system to find out who should take
// This can be done with one behaviour mode

// TODO: Convert the id into int datatype

// A note on convention before i forget:
// - Orders: Will be executed by elevator, will cause lights to activate
// - Requests: Abstract orders that haven't yet been confirmed/acknowledged

// TODO: Consider: Should obstruction be its own process?

func main() {
	// ---- Flags
	var port int
	var id string
	flag.IntVar(&port, "port", defaultElevatorPort, "Elevator server port")
	flag.StringVar(&id, "id", "", "Network node id")
	fmt.Println("Started!")

	flag.Parse()

	// ---- Initialize elevator
	elevio.Init("localhost:"+strconv.Itoa(port), elevalgo.NumFloors)
	elevalgo.InitFsm()

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)

	// ---- Initialize timer
	timer.SetTimeout(elevalgo.GetTimeout())
	timer.StartTimer()

	// ---- Initialize networking
	orderChan := make(chan requests.RequestInfo)
	peerRequestChan := make(chan requests.PeerRequest) // Node <- Network
	heartbeatChan := make(chan p2p.Heartbeat)
	elevatorStateChan := make(chan elevalgo.Elevator)

	go broadcast.BroadcastSender(stateBroadcastPort, heartbeatChan)
	go broadcast.BroadcastReceiver(stateBroadcastPort, heartbeatChan)

	go broadcast.BroadcastSender(p2p.RequestBroadCastPort, peerRequestChan)
	go broadcast.BroadcastReceiver(p2p.RequestBroadCastPort, peerRequestChan)

	go p2p.NodeProcess(heartbeatChan, peerRequestChan, drv_buttons, elevatorStateChan, orderChan, id)

	for {
		select {
		case requestInfo := <-orderChan:
			elevalgo.RequestButtonPressed(requestInfo.Floor, requestInfo.ButtonType)
		case floor := <-drv_floors:
			elevalgo.OnFloorArrival(floor)
		case obstructionEvent := <-drv_obstr:
			elevalgo.DoorObstructed(obstructionEvent)
		case <-timer.TimeoutChan:
			timer.StopTimer()
			elevalgo.OnDoorTimeout()
		default:
			elevatorStateChan <- elevalgo.GetState()
			timer.CheckTimeout()
		}
	}
}

// func main() {

// 	drv_buttons := make(chan elevio.ButtonEvent)
// 	drv_floors := make(chan int)
// 	drv_obstr := make(chan bool)
// 	poll_timer := make(chan bool, 1)
// 	requests := make(chan p2p.RequestInfo, RequestBufferSize)

// 	go elevio.PollButtons(drv_buttons)
// 	go elevio.PollFloorSensor(drv_floors)
// 	go elevio.PollObstructionSwitch(drv_obstr)

// 		case request := <-requests: // WARNING: sending things into this channel *will*
// 			// make the elevator service the request!
// 			fmt.Println("Received request")
// 			elevalgo.RequestButtonPressed(request.Floor, request.ButtonType)
// 		case button := <-drv_buttons:
// 			fmt.Println("Button press incoming!")
// 			request := p2p.RequestInfo{
// 				SenderId:   nodeInstance.GetId(),
// 				ButtonType: button.Button,
// 				Floor:      button.Floor,
// 			}

// 			if button.Button == elevio.BT_Cab {
// 				fmt.Println("Assigning to self")
// 				nodeInstance.AssignRequestSelf(request)
// 			} else {
// 				fmt.Println("Assigning to other")
// 				// nodeInstance.AssignRequest(request)
// 			}
// 		case floor := <-drv_floors:
// 			fmt.Println("Arrived on floor")
// 			elevalgo.OnFloorArrival(floor)
// 		case obstructed := <-drv_obstr:
// 			if obstructed {
// 				elevalgo.DoorObstructed()
// 			}
// 		case <-poll_timer:
// 			fmt.Println("Timer")
// 			timer.StopTimer()
// 			elevalgo.OnDoorTimeout()
// 		case request := <-nodeInstance.RequestChan:
// 			fmt.Println("Request arrived")
// 			if request.AssigneeID == nodeInstance.GetId() {
// 				nodeInstance.AssignRequestSelf(request.Request)
// 			}
// 			// default:
// 			// fmt.Println("Update started")

// 			// lightsState := elevalgo.MergeHallLights(
// 			// 	elevalgo.ThisElevator,
// 			// 	append(utils.MapToArray((nodeInstance.ExtractPeerState())), elevalgo.ThisElevator),
// 			// )
// 			// elevalgo.ThisElevator.SetLights(lightsState)

// 			// for _, outRequest := range nodeInstance.RequestsForPeers {
// 			// 	fmt.Println("Request for someone else:", outRequest)
// 			// 	nodeInstance.RequestChan <- outRequest
// 			// 	fmt.Println("Request sent")
// 			// }
// 			// time.Sleep(time.Millisecond * 10)
// 		}
// 	}
// }
