package networking

import elevalgo "sanntidslab/elevalgo"

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
