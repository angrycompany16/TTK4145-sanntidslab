package requests

import (
	"github.com/angrycompany16/driver-go/elevio"
)

// A type of request that is waiting to be received by some peer
type PeerRequest struct {
	// RequestInfo RequestInfo
	AssigneeID string
}

func NewPeerRequest(assigneeID string) PeerRequest {
	return PeerRequest{
		// RequestInfo: req,
		AssigneeID: assigneeID,
	}
}

// A type of request that is waiting to be acked by peers
type PendingRequest struct {
	// RequestInfo RequestInfo
	Active bool            // Is there a request here?
	Acks   map[string]bool // Array of booleans indicating whether the peers have accepted this request (i.e. whether they have backed it up or not)
}

type RequestInfo struct {
	ButtonType elevio.ButtonType
	Floor      int
}

func NewRequestInfo(buttonEvent elevio.ButtonEvent) RequestInfo {
	return RequestInfo{
		ButtonType: buttonEvent.Button,
		Floor:      buttonEvent.Floor,
	}
}
