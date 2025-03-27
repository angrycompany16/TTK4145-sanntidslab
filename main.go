package main

import (
	"flag"
	"fmt"
	"log"
	"path"
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
	obstructionTimeout  = time.Second * 4
	motorTimeout        = time.Second * 4
)

var configPath = path.Join("elevalgo", "config.yaml")

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
	config, err := elevalgo.LoadConfig(configPath)
	if err != nil {
		log.Fatal("Loading config failed with error", err)
	}

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
	go timer.RunTimer(resetDoorTimerChan, stopDoorTimerChan, doorTimeoutChan, config.DoorOpenDuration, false, "Door timer")

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

	// ---- Spawn core threads: networking, elevator, door and lights ----
	go networking.RunNode(
		buttonEventChan,
		nodeElevatorStateChan,
		orderChan,
		peerStateChan,
		config,
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
		config,
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
