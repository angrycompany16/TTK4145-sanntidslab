package main

import (
	"flag"
	"fmt"
	elevalgo "sanntidslab/elev_al_go"
	"sanntidslab/networking"
	"sanntidslab/timer"
	"strconv"
	"time"

	"github.com/angrycompany16/driver-go/elevio"
)

const (
	defaultElevatorPort = 15657
	obstructionTimeout  = time.Second * 3
	motorTimeout        = time.Second * 1000000
)

// TODO: A lot of things need more testing
// TODO: implement request assignment so it's actually correct
// - Doesn't work???
// TODO: Implement crashing in case of obstruction or motor failure
// TODO: Disconnect simulator: block all networking channels
// TODO: Lights
// TODO: Unit tests?

// TODO: *Read* code complete checklist properly and at least try to make the code
// quality good
// TODO: Consider: Should obstruction be its own process?

// TODO: Create publisher system to allow one channel to have multiple listeners?

// TODO: problem: Lights flicker a lot

func main() {
	// ---- Flags ----
	var port int
	var id string
	flag.IntVar(&port, "port", defaultElevatorPort, "Elevator server port")
	flag.StringVar(&id, "id", "", "Network node id")
	fmt.Println("Started!")

	flag.Parse()

	// ---- Initialize elevator and driver ----
	elevio.Init("localhost:"+strconv.Itoa(port), elevalgo.NumFloors)
	elevalgo.InitFsm()
	elevalgo.InitBetweenFloors()

	buttonEventChan := make(chan elevio.ButtonEvent, 1)
	floorChan := make(chan int)
	obstructionChan := make(chan bool)

	go elevio.PollButtons(buttonEventChan)
	// TODO: Continuous floor sensor detection
	go elevio.PollFloorSensor(floorChan)
	go elevio.PollObstructionSwitch(obstructionChan)

	// ---- Initialize timers ----
	// Door timer
	startDoorTimerChan := make(chan int)
	doorTimeoutChan := make(chan int)
	go timer.RunTimer(startDoorTimerChan, doorTimeoutChan, elevalgo.GetTimeout(), false, true, "Door timer")

	// Obstruction timer
	startObstructionTimerChan := make(chan int)
	obstructionTimeoutChan := make(chan int)
	go timer.RunTimer(startObstructionTimerChan, obstructionTimeoutChan, obstructionTimeout, true, true, "Obstruction timer")

	// Motor timer
	startMotorTimerChan := make(chan int)
	motorTimeoutChan := make(chan int)
	go timer.RunTimer(startMotorTimerChan, motorTimeoutChan, motorTimeout, true, false, "Motor timer")

	// ---- Communication with networking node ----
	orderChan := make(chan elevio.ButtonEvent, 1)
	elevatorStateChan := make(chan elevalgo.Elevator, 1)
	peerStateChan := make(chan []elevalgo.Elevator, 1)

	// ---- Spawn core threads, networking and elevator ----
	go networking.RunNode(
		buttonEventChan,
		elevatorStateChan,
		orderChan,
		peerStateChan,
		elevalgo.GetState(),
		id,
	)
	go elevalgo.RunElevator(
		floorChan,
		obstructionChan,
		orderChan,
		doorTimeoutChan,
		peerStateChan,
		elevatorStateChan,
		startDoorTimerChan,
		startObstructionTimerChan,
		startMotorTimerChan,
	)

	for {
		time.Sleep(1 * time.Second)
	}
}
