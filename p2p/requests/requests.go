package requests

import (
	"github.com/angrycompany16/driver-go/elevio"
)

// type Request interface {
// 	getRequestInfo() RequestInfo
// }

// A type of request that is waiting to be received by some peer
type PeerRequest struct {
	// RequestInfo RequestInfo
	AssigneeID string
}

// func (p *PeerRequest) getRequestInfo() RequestInfo {
// 	return p.RequestInfo
// }

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

// func (p *PendingRequest) getRequestInfo() RequestInfo {
// 	return p.RequestInfo
// }

func NewPendingRequest() PendingRequest {
	return PendingRequest{
		// RequestInfo: req,
		Acks: make(map[string]bool),
	}
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

// func ExtractPeerRequestInfo(input []PeerRequest) []RequestInfo {
// 	return utils.MapSlice(
// 		input,
// 		func(req PeerRequest) RequestInfo { return req.RequestInfo },
// 	)
// }

// func ExtractPendingRequestInfo(input []PendingRequest) []RequestInfo {
// 	return utils.MapSlice(
// 		input,
// 		func(req PendingRequest) RequestInfo { return req.RequestInfo },
// 	)
// }

// func RequestAlreadyExists(requestList []RequestInfo, request RequestInfo) bool {
// 	return slices.Contains(requestList, request)
// }
