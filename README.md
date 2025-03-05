Notes for the peer review group:

We have three main packages to look at in this module. Those are:
 - elev_algo: The main elevator. This package runs the elevator algorithm, and the only really interesting things to look at are the request assigner and maybe the file reading? (we used .yaml files as elevator config)
 - backup: This contains the code (very WIP) for our primary backup system. The idea here is that each elevator will back up a request before taking it and turning on the corresponding light, and that the backup acts as a service guarantee for hall calls.
 - network: This contains the code for nodes on the network. We have a peer discovery system with life signals, as well as connections between all the peers on the network. There is also some *very* barebones functionality for receiving hall calls to nodes.

The Network-go module is also in the handin (in a different folder). This is included because although we use UDP broadcast, we also use QUIC for p2p connections (https://en.wikipedia.org/wiki/QUIC). This is essentially a replacement for TCP connections. 