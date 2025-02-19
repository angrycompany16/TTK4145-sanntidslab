package main

import (
	"fmt"
	"os/exec"
)

var (
	ipaddress string = "10.100.23.23"
	password  string = "Sanntid15"
)

func runTerminalSSH(ipaddress string, password string) {

	// needs to source because we are SSHing into the remote machine
	_source := "source ~/.bashrc; export GO111MODULE=on; export GOROOT=/usr/local/go; export GOPATH=~/go;"

	// This is the command that will be run on the remote machine
	_runFile := _source + "cd ~/Documents/gruppe56/TTK4145-sanntidslab/ && go run main.go --shitNpiss; exec bash"

	// Makes sure that the terminal is detached from the current process (needs special characters to make it working)
	_commands := fmt.Sprintf("export DISPLAY=:0; nohup gnome-terminal -- bash -c \"%s\" > /dev/null 2>&1 &", _runFile)

	// SSH into the remote machine
	_ssh := fmt.Sprintf("sshpass -p '%s' ssh student@%s '%s'", password, ipaddress, _commands)

	// Build the full command and execute it
	_terminal := exec.Command("gnome-terminal", "--", "bash", "-c", _ssh)

	// smol error handling
	err := _terminal.Run()
	if err != nil {
		fmt.Println("Failed to run terminal%v", err)
	}
}

func runElevatorServer() { // refer to the runTerminalSSH function for comments
	_runFile := "elevatorserver; exec bash"

	_commands := fmt.Sprintf("export DISPLAY=:0; nohup gnome-terminal -- bash -c \"%s\" > /dev/null 2>&1 &", _runFile)

	_ssh := fmt.Sprintf("sshpass -p '%s' ssh student@%s '%s'", password, ipaddress, _commands)

	_terminal := exec.Command("gnome-terminal", "--", "bash", "-c", _ssh)

	err := _terminal.Run()
	if err != nil {
		fmt.Println("Failed to run elevatorserver %v", err)
	}
}

func main() {
	runElevatorServer()
	runTerminalSSH(ipaddress, password)
}
