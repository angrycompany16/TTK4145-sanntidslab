package main

import (
	"flag"
	"fmt"
	"sanntidslab/door"
	"sanntidslab/elevalgo"
	"sanntidslab/elevio"
	"sanntidslab/lights"
	"sanntidslab/networking"
	"sanntidslab/timer"
	"strconv"
	"time"
)

const (
	defaultElevatorPort = 15657
	obstructionTimeout  = time.Second * 10
	motorTimeout        = time.Second * 10
)

// TODO: Final todo list before FAT:
// - Convert door into its own process ✓
// - Fully implement obstruction switch and motor blockage timers ✓
// - Change elevator state sending from continuous to diff ✓
// - Change peer list sending from continuous to diff ✓
// - Implement virtual state for pending requests ✓
// - Fix the request assigner
// - Do a *lot* more testing with packet loss, both working and not working
// - Make the hardwareCommands solution cleaner [?]
// - Gather everything into one repo ✓
// - Review the structure of elevalgo ✓
// - Read the project spec completely, and verify that everything works
// - Do test FAT

// A problem with packet loss:
// -----------------------------
// If there is high packet loss and an elevator disconnects, the other two elevators
// may take requests which haven't been fully acked. This can happen if elevator 1
// takes a request while elevator 3 is "disconnected", elevator 1 then detects that
// only elevator 2 needs to ack, which happens, and then elevator 1 takes the request
// with only an ack from elevator 2. If elevator 2 then dies, the request has not been
// backed up
// One reason why this may potentially not be such a huge problem is that
// each elevator broadcasts its state either way, so there is a large chance
// that elevator 3 will pick up that elevator 1 is taking the request anyways, and then
// if elevator 1 dies, elevator 3 can take over / back up the request

// But in general packet loss will ravage our elevator system
// :DD
func main() {
	// ---- Flags ----
	var port int
	var id string
	flag.IntVar(&port, "port", defaultElevatorPort, "Elevator server port")
	flag.StringVar(&id, "id", "", "Network node id")
	fmt.Println("Started!")

	flag.Parse()

	// // ---- Initialize elevator ----
	elevio.Init("localhost:"+strconv.Itoa(port), elevalgo.NumFloors)
	elevalgo.InitFsm()
	initElevator, doorOpenDuration := elevalgo.InitBetweenFloors()

	// ---- Initialize hardware communication ----
	buttonEventChan := make(chan elevio.ButtonEvent, 1)
	floorChan := make(chan int)
	obstructionChan := make(chan bool)

	go elevio.PollButtons(buttonEventChan)
	go elevio.PollFloorSensor(floorChan)
	go elevio.PollObstructionSwitch(obstructionChan)

	obstructionInit := <-obstructionChan

	// ---- Initialize timers ----
	// Door timer
	resetDoorTimerChan := make(chan int)
	stopDoorTimerChan := make(chan int)
	doorTimeoutChan := make(chan int)
	go timer.RunTimer(resetDoorTimerChan, stopDoorTimerChan, doorTimeoutChan, doorOpenDuration, false, "Door timer")

	// Obstruction timer
	resetObstructionTimerChan := make(chan int)
	stopObstructionTimerChan := make(chan int)
	obstructionTimeoutChan := make(chan int)
	go timer.RunTimer(resetObstructionTimerChan, stopObstructionTimerChan, obstructionTimeoutChan, obstructionTimeout, true, "Obstruction timer")

	// Motor timer
	resetMotorTimerChan := make(chan int)
	stopMotorTimerChan := make(chan int)
	motorTimeoutChan := make(chan int)
	go timer.RunTimer(resetMotorTimerChan, stopMotorTimerChan, motorTimeoutChan, motorTimeout, true, "Motor timer")

	// ---- Networking node communication ----
	// TODO: Try unbuffering some of these channels and see what happens
	orderChan := make(chan elevio.ButtonEvent, 1)
	nodeElevatorStateChan := make(chan elevalgo.Elevator, 1)
	peerStateChan := make(chan []elevalgo.Elevator, 1)

	// ---- Door communication ----
	doorRequestChan := make(chan int)
	doorCloseChan := make(chan int)

	// ---- Lights communication
	lightsElevatorStateChan := make(chan elevalgo.Elevator, 1)

	// ---- Spawn core threads: networking, elevator, door and lights ----
	go networking.RunNode(
		buttonEventChan,
		nodeElevatorStateChan,
		orderChan,
		peerStateChan,
		initElevator,
		id,
	)

	go elevalgo.RunElevator(
		floorChan,
		orderChan,
		doorCloseChan,
		doorRequestChan,
		lightsElevatorStateChan,
		nodeElevatorStateChan,
		resetMotorTimerChan,
		stopMotorTimerChan,
	)

	go door.RunDoor(
		obstructionChan,
		doorTimeoutChan,
		doorRequestChan,
		doorCloseChan,
		resetDoorTimerChan,
		resetObstructionTimerChan,
		stopObstructionTimerChan,
		obstructionInit,
	)

	go lights.RunLights(lightsElevatorStateChan, peerStateChan)

	for {
		time.Sleep(time.Second)
	}
}
