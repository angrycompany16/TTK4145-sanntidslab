#!/bin/bash


go build -o supervisor_go supervisor/supervisor.go
go build -o elevator_go main.go

bash -c 'pkill supervisor_go'
bash -c 'pkill elevator_go'
bash -c 'pkill elevatorserver'
bash -c 'pkill -f SimElevatorServer'

if [ -z "$2" ]; then
    gnome-terminal -- bash -c "elevatorserver; exec bash"
    gnome-terminal -- bash -c "./elevator_go -id $1; exec bash"
    gnome-terminal -- bash -c "./supervisor_go -id $1; exec bash"
else
    gnome-terminal -- bash -c "./SimElevatorServer; exec bash"
    gnome-terminal -- bash -c "./elevator_go -id $1 -port $2; exec bash"
    gnome-terminal -- bash -c "./supervisor_go -id $1 -port $2 ; exec bash"
fi


exit 0