package main

import (
	"fmt"
	"os/exec"
	"time"
)

var (
	timeout      time.Duration = 20 * time.Millisecond
	path         string        = "Documents/gruppe56/TTK4145-sanntidslab/"
	elevatorFlag string        = "ElevatorNode"
	// Backup                     = localBackup{}
)

// type localBackup struct {
// 	localState     elevalgo.Elevator
// 	localWorldView map[string]PeerInfo
// 	localTimeAlive time.Time // maybe ticker?
// }

func InitProcessPair() {
	fmt.Printf("Initializing process pair")
	runFile := fmt.Sprintf("cd ~%s && go run .; exec bash", path)
	exec.Command("gnome-terminal", "--", "bash", "-c", runFile).Run()
}

func main() {

	reviving := false
	aliveChan := make(chan int)

	go processIsAlive(elevatorFlag, aliveChan)

	for {
		select {

		case msg := <-aliveChan:
			reviving = tryRevive(msg, reviving)

			// default: // not sure waht we want to to, either local elevator is readable and i can copy, or i have to listen to some chan?
			// 	Backup.localState = localElevator.State
			// 	Backup.localWorldView = localElevator.WorldView
			// 	Backup.localTimeAlive = localElevator.timeAlive
		}
	}
}

func tryRevive(msg int, reviveFlag bool) (reviving bool) {

	if msg != 0 && !reviveFlag {
		fmt.Printf("tried to revive, recieved %d, %t \n", msg, reviveFlag)
		//reviveElevator()
		reviveFlag = true
	}

	if msg == 0 && reviveFlag {
		// send info on a channel & block untill elevator answers or initialize elevator with a state???
		reviveFlag = false
	}

	return reviveFlag
}

func processIsAlive(flag string, ch chan int) {
	terminal := exec.Command("pgrep -fl ", flag)

	for {
		if err := terminal.Run(); err != nil {
			panic(err)
			if exitError, ok := err.(*exec.ExitError); ok {
				fmt.Printf("exitError : %d", exitError.ExitCode())
				ch <- exitError.ExitCode()
			}
			fmt.Printf("didnt find")
			ch <- 1
		}
		ch <- 0
		time.Sleep(timeout)
	}
}

func reviveElevator() {
	fmt.Printf("Running Elevator\n")
	runFile := fmt.Sprintf("cd ~%s && go run .; exec bash", path)
	exec.Command("gnome-terminal", "--", "bash", "-c", runFile).Run()
}
