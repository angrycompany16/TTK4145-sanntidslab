#!/bin/bash

vars=$(getopt -o i:p:v --long id:,port:,verbose -- "$@")
eval set -- "$vars"

# Flags: i, p, v
id=''
port=''
exec_bash=''
verbose=''

go build -o bin/supervisor_go supervisor/supervisor.go
go build -o bin/elevator_go main.go

bash -c 'pkill supervisor_go'
bash -c 'pkill elevator_go'
bash -c 'pkill elevatorserver'
bash -c 'pkill -f SimElevatorServer'

# the command exec bash makes the process and terminal independent. If we want terminals to close when killed remove it.

for opt; do
    case "$opt" in
        -i|--id) id="$2"; shift 2 ;;
        -p|--port) port="$2"; shift 2 ;;
        -v|--verbose) exec_bash="exec bash"; verbose="true" echo "Hello"; shift ;;
    esac
done

if [ -z "$port" ]; then
    echo "Running without port"
    gnome-terminal -- bash -c "elevatorserver; $exec_bash"
    gnome-terminal -- bash -c "./bin/elevator_go -id $id; $exec_bash"
    gnome-terminal -- bash -c "./bin/supervisor_go -id $id -verbose $verbose; $exec_bash"
else
    echo "Running with port"
    gnome-terminal -- bash -c "./bin/SimElevatorServer --port $port; $exec_bash"
    gnome-terminal -- bash -c "./bin/elevator_go -id $id -port $port; $exec_bash"
    gnome-terminal -- bash -c "./bin/supervisor_go -id $id -port $port -verbose $verbose; $exec_bash"
fi


exit 0