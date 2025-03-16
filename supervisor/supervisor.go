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
	pwd  string
	id   string
	port int
)

func main() {
	pwd = getArgs()

	reviving := false
	aliveChan := make(chan int)

	go processIsAlive(elevatorFlag, aliveChan)

	for msg := range aliveChan {
		reviving = tryRevive(msg, reviving)
	}
}

func getArgs() (pwd string) {

	flag.IntVar(&port, "port", defaultElevatorPort, "Elevator server port")
	flag.StringVar(&id, "id", "", "Network node id")
	flag.Parse()

	pwd, _ = os.Getwd()
	return pwd
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

	if port == defaultElevatorPort {
		runFile := fmt.Sprintf("cd %s && ./run.sh %s; exec bash", pwd, id)
		exec.Command("gnome-terminal", "--", "bash", "-c", runFile).Run()
		return
	}

	runFile := fmt.Sprintf("cd %s && ./run.sh %s %d; exec bash", pwd, id, port)
	exec.Command("gnome-terminal", "--", "bash", "-c", runFile).Run()
}
