# Lab work group -- TTK4145

In this lab project, we are implementing a distributed elevator system with n elevators, with a strong emphasis on fault tolerance.

The project can be run using the included shell script. This automatically builds the program, starts either the elevatorServer or the SimElevatorServer together with our elevator program and its supervisor.

```./run.sh --id X --port Y --verbose (optional argument with no value) ```

Ensure that the different elevators are spawned with unique IDs.

If a port is provided, the simulator will be run, located in simulator/SimElevatorServer. 

If no port is specified, the system will default to port `15657` and run the physical elevator.