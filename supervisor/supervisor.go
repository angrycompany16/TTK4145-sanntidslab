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
	timeout             time.Duration = 5000 * time.Millisecond
	defaultElevatorPort               = 15657
)

var (
	pwd, _ = os.Getwd()
	id     string
	port   int
)

func main() {
	flag.IntVar(&port, "port", defaultElevatorPort, "Elevator server port")
	flag.StringVar(&id, "id", "", "Network node id")
	flag.Parse()

	reviving := false
	aliveChan := make(chan int)

	go processIsAlive(elevatorFlag, aliveChan)

	for {
		select {

		case msg := <-aliveChan:
			reviving = tryRevive(msg, reviving)

		}
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
	if port == defaultElevatorPort {
		runFile := fmt.Sprintf("cd %s && ./run.sh %s; exec bash", pwd, id)
		exec.Command("gnome-terminal", "--", "bash", "-c", runFile).Run()
		return
	}
	runFile := fmt.Sprintf("cd %s && ./run.sh %s %d; exec bash", pwd, id, port)
	exec.Command("gnome-terminal", "--", "bash", "-c", runFile).Run()
}
