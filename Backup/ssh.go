package backup

import (
	"errors"
	"fmt"
	"os/exec"
)

// Queries the remote machine and returns 1 if the process doesnt't exist, and 0 if it does
func AlreadyRunning(processName string, ipaddress string, password string) (exitCode int, err error) {
	var _terminal *exec.Cmd

	if ipaddress == "localhost" {
		_runFile := "pgrep -fl " + processName
		_terminal = exec.Command("bash", "-c", _runFile)

	} else {
		_runFile := "pgrep -fl " + processName
		_ssh := fmt.Sprintf("sshpass -p '%s' ssh student@%s '%s'", password, ipaddress, _runFile)
		_terminal = exec.Command("bash", "-c", _ssh)
	}

	if err := _terminal.Run(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode(), nil
		}
		return 0, errors.New("failed to run terminal")
	}
	return 0, nil
}

// SSH into the remote machine and create backup instance
func CreateBackupListener(ipaddress string, hostIP string, password string) (created bool) {
	fmt.Print("Creating Backup")

	// Need to source to be able to run go
	_source := "source ~/.bashrc; export GO111MODULE=on; export GOROOT=/usr/local/go; export GOPATH=~/go;"

	// This is the command that will be run on the remote machine
	_runFile := _source + fmt.Sprintf(
		"cd ~/Documents/gruppe56/TTK4145-sanntidslab/backup && go run backup_listener.go node=%s hostIP=%s; exec bash", backupFlag, hostIP)

	// Makes sure that the terminal is detached from the current process (needs special characters to make it working)
	_commands := fmt.Sprintf("export DISPLAY=:0; nohup gnome-terminal -- bash -c \"%s\" > /dev/null 2>&1 &", _runFile)

	// SSH into the remote machine
	_ssh := fmt.Sprintf("sshpass -p '%s' ssh student@%s '%s'", password, ipaddress, _commands)

	// Build the full command and execute
	_terminal := exec.Command("gnome-terminal", "--", "bash", "-c", _ssh)

	// smol error handling
	err := _terminal.Run()
	if err != nil {
		fmt.Printf("Failed to run terminal%v", err)
		created = false
	} else {
		created = true
	}

	return created
}

// If fail try to remove host key, and generate a new one manually through SSH
// needs testing, might be more convinient to generate a key once, and have it on all
//  sanntidlab machines we intend to use

func ReviveElevator(ipaddress string, password string) {

	fmt.Print("Running Elevator\n")

	_source := "source ~/.bashrc; export GO111MODULE=on; export GOROOT=/usr/local/go; export GOPATH=~/go;"
	_runFile := _source + "cd ~/Documents/gruppe56/TTK4145-sanntidslab/ && go run . --mode=normal --node=primary; exec bash"

	_commands := fmt.Sprintf("export DISPLAY=:0; nohup gnome-terminal -- bash -c \"%s\" > /dev/null 2>&1 &", _runFile)
	_ssh := fmt.Sprintf("sshpass -p '%s' ssh student@%s '%s'", password, ipaddress, _commands)

	_terminal := exec.Command("gnome-terminal", "--", "bash", "-c", _ssh)
	err := _terminal.Run()
	if err != nil {
		fmt.Printf("Failed to run terminal%v", err)
	}
}

func ReviveElevatorserver(ipaddress string, password string) { // refer to the runTerminalSSH function for comments

	fmt.Print("Running ElevatorServer\n")

	_runFile := "elevatorserver; exec bash"

	_commands := fmt.Sprintf("export DISPLAY=:0; nohup gnome-terminal -- bash -c \"%s\" > /dev/null 2>&1 &", _runFile)
	_ssh := fmt.Sprintf("sshpass -p '%s' ssh student@%s '%s'", password, ipaddress, _commands)

	_terminal := exec.Command("gnome-terminal", "--", "bash", "-c", _ssh)
	err := _terminal.Run()
	if err != nil {
		fmt.Printf("Failed to run elevatorserver %v", err)
	}
}

func Revive(ipaddress string, password string) [2]int {
	var exitCodes [2]int

	// Check if the elevatorserver is already running, and if not start it up
	exitCode, err := AlreadyRunning("elevatorserver", ipaddress, password)
	if err != nil {
		fmt.Printf("Failed to Query %v", err)

	} else if exitCode == 0 {
		fmt.Println("Elevatorserver is already running, do nothing")
		exitCodes[0] = 0

	} else {
		fmt.Println("Elevatorserver is not running")
		ReviveElevatorserver(ipaddress, password)
		exitCodes[0] = 1
	}

	// Check if the primary (elevator) is already running, and if not start it up
	exitCode, err = AlreadyRunning("primary", ipaddress, password)
	if err != nil {
		fmt.Printf("Failed to Query %v", err)

	} else if exitCode == 0 {
		fmt.Println("Primary is already running, do nothing")
		exitCodes[1] = 0

	} else {
		fmt.Println("Script is not running")
		ReviveElevator(ipaddress, password)
		exitCodes[1] = 1
	}

	return exitCodes // use codes to figure out if somebody was revived.
}