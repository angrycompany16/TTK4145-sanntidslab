# Automatically builds and runs the elevator program, elevator server (or simulator) and
# supervisor based on the given flags. Accepts an id, port, and verbose (whether 
# terminals should close automatically or not)
# Note that running without a port argument will start the elevatorserver rather than
# the simulator.

#!/bin/bash
vars=$(getopt -o i:p:v --long id:,port:,verbose -- "$@")
eval set -- "$vars"

id=''
port=''
exec_bash=''
verbose=''

supervisor_exe='supervisor_go'
elevator_exe='elevator_go'

go build -o bin/$supervisor_exe supervisor/supervisor.go
go build -o bin/$elevator_exe main.go

bash -c "pkill $supervisor_exe"
bash -c "pkill $elevator_exe"
bash -c 'pkill elevatorserver'
bash -c 'pkill -f SimElevatorServer'

for opt; do
    case "$opt" in
        -i|--id) id="$2"; shift 2 ;;
        -p|--port) port="$2"; shift 2 ;;
        -v|--verbose) exec_bash="exec bash"; verbose="true" echo "Hello"; shift ;;
    esac
done

if [ -z "$port" ]; then
    gnome-terminal -- bash -c "elevatorserver; $exec_bash"
    gnome-terminal -- bash -c "./bin/$elevator_exe -id $id; $exec_bash"
    gnome-terminal -- bash -c "./bin/$supervisor_exe -id $id -verbose $verbose; $exec_bash"
else
    gnome-terminal -- bash -c "./simulator/SimElevatorServer --port $port; $exec_bash"
    gnome-terminal -- bash -c "./bin/$elevator_exe -id $id -port $port; $exec_bash"
    gnome-terminal -- bash -c "./bin/$supervisor_exe -id $id -port $port -verbose $verbose; $exec_bash"
fi

exit 0