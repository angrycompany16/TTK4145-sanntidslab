This is how the networking module will be implemented

Preliminary questions:
- What if one node loses connection?
	- Suggestion: Try to ping the node until it can connect again
- What if a node loses power briefly?
	- Suggestion: Restart the node somehow (?)
- What if some unforeseen event causes the elevator to never reach its destination, but communication remains intact?
	- Somehow get error message, make some other node control the elevator
- Do all nodes need to agree about a call for it to be accepted?
	- No, just one node needs to. Calls are sent out to every node in the network
* How can you be sure that a remote node "agrees" on a call?
	* What is meant by this
* How do you handle packet loss?
	* Send packets out to each other node, or just use TCP lol
* Do you share the entire state of the current calls, or just 

Let's do this shit in TCP
Peer to peer is probably best
Make every elevator node initialize a connection to every other node
Should be able to pass structs between peers

Peers have a timeout, where if they do nothing after the timeout they get checked on by other peers.