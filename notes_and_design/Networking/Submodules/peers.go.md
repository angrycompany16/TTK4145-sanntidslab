This module handles the detection of inactive peers in the networking module. Functions similarly to [[bcast.go]], but it is specialized for sending and handling `PeerUpdate` messages.

```go
type PeerUpdate struct {
	Peers []string
	New string
	Lost []string
}
```
A message that is passed every `interval`, which informs about the state of the peers in this moment, whether any peers have been created or not etc. `timeout` determines how long a peer can be inactive before it is considered "Lost".

```go
func Transmitter(port int, id string, transmitEnable <-chan bool)
```
A function which, whenever anything is receiver on the `transmitEnable` channel, broadcasts `id` on the specified port. This lets the other nodes that are listening know that the peer is still alive.


```go
func Receiver(port int, peerUpdateCh chan<- PeerUpdate)
```
This function reads from `port` (with read timeout of `interval`), and first checks whether any new peers are added to the network. Then it checks whether a peer has timed out (not signaled its existence within `timeout`), and lastly organizes and sends a `PeerUpdate` to `peerUpdateCh` containing the new `PeerUpdate`.

The usage pattern looks like this:
```go
// Setup
send_channel := make(chan some_datatype)
receive_channel := make(chan some_datatype)
/* id is a unique string for each peer */
go peers.Transmitter(port_send, id, peerTxEnable)
go peers.Receiver(port_recv, peerUpdateCh)


// Usage
for {
	select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf(" Peers: %q\n", p.Peers)
			fmt.Printf(" New: %q\n", p.New)
			fmt.Printf(" Lost: %q\n", p.Lost)
	}
}
```
