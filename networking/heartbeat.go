package networking

import elevalgo "sanntidslab/elevalgo"

// Contains all information needed to ensure consistency among peers, should always be
// sent periodically
type Heartbeat struct {
	SenderId        string
	Uptime          int64
	State           elevalgo.Elevator
	WorldView       map[string]peer
	PendingRequests PendingRequestList
}

func newHeartbeat(node node) Heartbeat {
	return Heartbeat{
		SenderId:        nodeID,
		Uptime:          uptime,
		State:           node.state,
		PendingRequests: node.pendingRequestList,
		WorldView:       node.peers,
	}
}
