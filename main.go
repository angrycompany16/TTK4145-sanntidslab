package main

import (
	"flag"
	"fmt"
	elevalgo "sanntidslab/elev_al_go"
	"sanntidslab/elev_al_go/timer"
	"sanntidslab/peer"
	"strconv"
	"time"

	"github.com/angrycompany16/driver-go/elevio"
)

const (
	requestBufferSize   = 1
	defaultElevatorPort = 15657 /* I think? */

)

// TODO: *Read* code complete checklist properly and at least try to make the code
// quality good

// A note on convention before i forget:
// - Orders: Will be executed by elevator, will cause lights to activate
// - Requests: Abstract orders that haven't yet been confirmed/acknowledged

// TODO: Consider: Should obstruction be its own process?

// TODO: Implement crashing in case of obstruction or motor failure

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
	elevalgo.InitBetweenFloors()

	buttonEventChan := make(chan elevio.ButtonEvent, 1)
	floorChan := make(chan int)
	obstructionChan := make(chan bool)

	go elevio.PollButtons(buttonEventChan) // "Sent" to node for further action

	go elevio.PollFloorSensor(floorChan)             // "Sent" to elevalgo for declaring new state
	go elevio.PollObstructionSwitch(obstructionChan) // "Sent" to elevalgo for declaring new state

	// ---- Initialize timer
	timer.SetTimeout(elevalgo.GetTimeout())
	timer.StartTimer()

	// ---- Initialize networking
	orderChan := make(chan elevio.ButtonEvent, 1)
	elevatorStateChan := make(chan elevalgo.Elevator)

	go peer.NodeProcess(buttonEventChan, elevatorStateChan, orderChan, elevalgo.GetState(), id)

	go elevalgo.ElevatorProcess(floorChan, obstructionChan, orderChan, elevatorStateChan)
	for {
		time.Sleep(1 * time.Second)
	}

}
