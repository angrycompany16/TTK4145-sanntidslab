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
	orderChan := make(chan elevio.ButtonEvent, 1)
	nodeElevatorStateChan := make(chan elevalgo.Elevator, 1)
	peerStateChan := make(chan []elevalgo.Elevator, 1)

	// ---- Door communication ----
	doorRequestChan := make(chan int)
	doorCloseChan := make(chan int)

	// ---- Lights communication
	lightsElevatorStateChan := make(chan elevalgo.Elevator, 1)

	// ---- Disconnect ----
	disconnectChan := make(chan int, 1)

	// ---- Spawn core threads: networking, elevator, door and lights ----
	go networking.RunNode(
		buttonEventChan,
		nodeElevatorStateChan,
		orderChan,
		peerStateChan,
		disconnectChan,
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
