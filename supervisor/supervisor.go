package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"
)

const (
	executableName      string        = "elevator_go"
	timeout             time.Duration = 2000 * time.Millisecond
	defaultElevatorPort               = 15657
)

var (
	pwd     string
	id      string
	port    int
	verbose bool
)

// Runs a supervisor to ensure elevators do not remain in a crashed state
func main() {
	flag.IntVar(&port, "port", defaultElevatorPort, "Elevator server port")
	flag.StringVar(&id, "id", "", "Network node id")
	flag.BoolVar(&verbose, "verbose", false, "Run with exec bash or not (debugging)")
	flag.Parse()

	pwd, _ = os.Getwd()

	for {
		err := exec.Command("pgrep", "-f", executableName).Run()

		if err == nil {
			fmt.Println("0")
		} else {
			fmt.Println("1")
			reviveElevator()
		}

		time.Sleep(timeout)
	}
}

func reviveElevator() {
	fmt.Println("Running run.sh")

	exec_bash := "; exec bash"
	if verbose {
		exec_bash = "--verbose; exec bash"
	}

	if port == defaultElevatorPort {
		runFile := fmt.Sprintf("cd %s && ./run.sh --id %s %s", pwd, id, exec_bash)
		exec.Command("gnome-terminal", "--", "bash", "-c", runFile).Run()
		return
	}

	runFile := fmt.Sprintf("cd %s && ./run.sh --id %s --port %d %s", pwd, id, port, exec_bash)
	fmt.Println(runFile)
	exec.Command("gnome-terminal", "--", "bash", "-c", runFile).Run()
}
