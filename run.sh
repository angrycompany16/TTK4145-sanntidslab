#!/bin/bash

go build -o supervisor_go supervisor/supervisor.go
go build -o elevator_go main.go

bash -c 'pkill supervisor_go'
bash -c 'pkill elevator_go'
bash -c 'pkill SimElevatorServer'

gnome-terminal -- bash -c "./SimElevatorServer; exec bash"
gnome-terminal -- bash -c "./elevator_go; exec bash"
gnome-terminal -- bash -c "./supervisor_go; exec bash"

exit 0