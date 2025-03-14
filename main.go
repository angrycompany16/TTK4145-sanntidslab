package main

import (
	"flag"
	"fmt"
	"internal/itoa"
	elevalgo "sanntidslab/elev_al_go"
	"sanntidslab/elev_al_go/timer"
	"sanntidslab/p2p"
	"sanntidslab/p2p/requests"

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
	var port, id int
	flag.IntVar(&port, "port", defaultElevatorPort, "Elevator server port")
	flag.IntVar(&id, "id", 0, "Network node id")
	fmt.Println("Started!")

	flag.Parse()

	// ---- Initialize elevator
	elevio.Init("localhost:"+itoa.Itoa(port), elevalgo.NumFloors)
	elevalgo.InitFsm()

	buttonEventChan := make(chan elevio.ButtonEvent)
	floorChan := make(chan int)
	obstructionChan := make(chan bool)

	go elevio.PollButtons(buttonEventChan) // "Sent" to node for further action

	go elevio.PollFloorSensor(floorChan) // "Sent" to elevalgo for declaring new state
	go elevio.PollObstructionSwitch(obstructionChan) // "Sent" to elevalgo for declaring new state

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

	go p2p.NodeProcess(heartbeatChan, peerRequestChan, buttonEventChan, elevatorStateChan, orderChan, id)

	go elevalgo.ElevatorProcess(floorChan, obstructionChan, orderChan, elevatorStateChan)
}


