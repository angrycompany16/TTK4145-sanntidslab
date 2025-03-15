package main

import (
	"flag"
	"fmt"
	elevalgo "sanntidslab/elev_al_go"
	"sanntidslab/elev_al_go/timer"
	"sanntidslab/peer"
	"sanntidslab/utils"
	"strconv"

	"github.com/angrycompany16/Network-go/network/transfer"
	"github.com/angrycompany16/driver-go/elevio"
)

const (
	requestBufferSize    = 1
	defaultElevatorPort  = 15657 /* I think? */
	stateBroadcastPort   = 36251 // Akkordrekke
	requestBroadCastPort = 12345
)

// TODO: *Read* code complete checklist properly and at least try to make the code
// quality good

// TODO: Implement the backup actually taking lost requests itself
// TODO: In case of disconnect, all requests should also be taken
// TODO: Arbitration/priority system to find out who should take
// This can be done with one behaviour mode

// A note on convention before i forget:
// - Orders: Will be executed by elevator, will cause lights to activate
// - Requests: Abstract orders that haven't yet been confirmed/acknowledged

// TODO: Consider: Should obstruction be its own process?

// TODO: Door not working

func main() {
	// ---- Flags
	var port int
	flag.IntVar(&port, "port", defaultElevatorPort, "Elevator server port")
	flag.StringVar(&peer.GlobalID, "id", "", "Network node id")
	fmt.Println("Started!")

	flag.Parse()

	// ---- Initialize elevator
	elevio.Init("localhost:"+strconv.Itoa(port), elevalgo.NumFloors)
	elevalgo.InitFsm()
	elevalgo.InitBetweenFloors()

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
	orderChan := make(chan elevio.ButtonEvent, 1)
	peerRequestChan := make(chan peer.Advertiser) // Node <- Network
	heartbeatChan := make(chan peer.Heartbeat)
	elevatorStateChan := make(chan elevalgo.Elevator)

	go transfer.BroadcastSender(stateBroadcastPort, heartbeatChan)
	go transfer.BroadcastReceiver(stateBroadcastPort, heartbeatChan)

	go transfer.BroadcastSender(requestBroadCastPort, peerRequestChan)
	go transfer.BroadcastReceiver(requestBroadCastPort, peerRequestChan)

	go p2p.NodeProcess(heartbeatChan, peerRequestChan, buttonEventChan, elevatorStateChan, orderChan, id)

	go elevalgo.ElevatorProcess(floorChan, obstructionChan, orderChan, elevatorStateChan)

}

