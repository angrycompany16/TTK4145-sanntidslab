package processpair

import (
	"fmt"
	"os/exec"
	"time"
)

var (
	timeout      time.Duration = 20 * time.Millisecond
	path         string        = "Documents/gruppe56/TTK4145-sanntidslab/"
	elevatorFlag string        = "ElevatorNode"
	Backup                     = localBackup{}
)

type localBackup struct {
	localState     elevalgo.Elevator
	localWorldView map[string]PeerInfo
	localTimeAlive time.Time // maybe ticker?
}

func InitProcessPair() {
	fmt.Printf("Initializing Backup")
	runFile := fmt.Sprintf("cd ~%s && go run .; exec bash")
	exec.Command("gnome-terminal", "--", "bash", "-c", runFile).Run()
}

func main() {

	reviving := false
	aliveChan := make(chan int)

	go processIsAlive(elevatorFlag, aliveChan)

	for {
		select {

		case msg := <-aliveChan:

			if msg != 0 && !reviving {
				reviveElevator()
				reviving = true
			}

			if msg == 0 && reviving {
				// send info on a channel & block untill elevator answers or initialize elevator with a state???
				reviving = false
			}

		default: // not sure waht we want to to, either local elevator is readable and i can copy, or i have to listen to some chan?
			Backup.localState = localElevator.State
			Backup.localWorldView = localElevator.WorldView
			Backup.localTimeAlive = localElevator.timeAlive
		}
	}
}

func processIsAlive(flag string, ch chan int) {
	terminal := exec.Command("bash", "-c", "pgrep -fl ", flag)

	for {
		if err := terminal.Run(); err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				ch <- exitError.ExitCode()
			}
			ch <- 1
		}
		ch <- 0
		time.Sleep(timeout)
	}
}

func reviveElevator() {
	fmt.Printf("Running Elevator\n")
	runFile := fmt.Sprintf("cd ~%s && go run .; exec bash")
	exec.Command("gnome-terminal", "--", "bash", "-c", runFile).Run()
}
