# Lab work group 56 TTK4145

In this lab project we are implementing a distributed elevator system with n elevators, with heavy emphasis on fault tolerance.

To run the project, make sure that either the elevatorserver or the simulator (https://github.com/TTK4145/Simulator-v2) is already running, and run the program with

```go run . -id X -port Y```

Make sure that the port flag corresponds to the port on which the elevator/simulator is running. If no port is specified it will choose the default port `15657`. Also ensure that the elevators are spawned with unique IDs.
