package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"
)

const (
	elevatorFlag        string        = "elevator_go"
	timeout             time.Duration = 2000 * time.Millisecond
	defaultElevatorPort               = 15657
)

var (
	pwd     string
	id      string
	port    int
	verbose bool
)

func main() {
	flag.IntVar(&port, "port", defaultElevatorPort, "Elevator server port")
	flag.StringVar(&id, "id", "", "Network node id")
	flag.BoolVar(&verbose, "verbose", false, "Run with exec bash or not (debugging)")
	flag.Parse()

	pwd, _ = os.Getwd()

	reviving := false
	aliveChan := make(chan int)

	go processIsAlive(elevatorFlag, aliveChan)

	for msg := range aliveChan {
		reviving = tryRevive(msg, reviving)
	}
}

func tryRevive(msg int, reviveFlag bool) (reviving bool) {

	if msg != 0 && !reviveFlag {
		fmt.Printf("tried to revive, recieved %d, %t \n", msg, reviveFlag)
		reviveElevator()
		time.Sleep(timeout)
		reviveFlag = true
	}

	if msg == 0 && reviveFlag {
		reviveFlag = false
	}

	return reviveFlag
}

func processIsAlive(flag string, aliveChan chan<- int) {

	for {
		err := exec.Command("pgrep", "-f", flag).Run()

		if err == nil {
			fmt.Println("0")
			aliveChan <- 0
		} else {
			fmt.Println("1")
			aliveChan <- 1
		}

		time.Sleep(timeout)
	}
}

func reviveElevator() {
	fmt.Println("Running run.sh")

	// ;exec bash makes it so that the terminal persists in spite of the process its running
	// being terminated, so if we do not want this simply remove the end part.

	exec_bash := ";"
	if verbose {
		exec_bash = "--verbose; exec bash"
	}

	if port == defaultElevatorPort {
		runFile := fmt.Sprintf("cd %s && ./run.sh --id %s %s", pwd, id, exec_bash)
		exec.Command("gnome-terminal", "--", "bash", "-c", runFile).Run()
		return
	}

	runFile := fmt.Sprintf("cd %s && ./run.sh --id %s --port %d %s", pwd, id, port, exec_bash)
	exec.Command("gnome-terminal", "--", "bash", "-c", runFile).Run()
}
