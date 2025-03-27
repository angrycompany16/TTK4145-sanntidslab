# Lab work group 56 TTK4145

In this lab project, we are implementing a distributed elevator system with n elevators, with a strong emphasis on fault tolerance.

The project can be run using the included shell script. This automatically starts either the elevatorServer or the SimElevatorServer together with our elevator program and its supervisor.

```./run.sh --id X --port Y --verbose Z ```

Ensure that the different elevators are spawned with unique IDs.

If a port is provided, the simulator will be run. Make sure that the executable simulator (https://github.com/TTK4145/Simulator-v2) is located in the "/bin" folder.

If no port is specified, the system will default to port `15657` and run the physical elevator.

