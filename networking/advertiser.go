package networking

import (
	"fmt"
	elevalgo "sanntidslab/elevalgo"
	"sanntidslab/elevio"

	"github.com/google/uuid"
)

// List of requests being advertised to other peers
type Advertiser struct {
	Requests [elevalgo.NumFloors][elevalgo.NumButtons]AdvertisedRequest
}

type AdvertisedRequest struct {
	AssigneeID string
	UUID       string // Used for differentiating advertised requests that have the same data
}

// Stops advertising if a peer is actively taking the request
func updateAdvertiser(_node node) Advertiser {
	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			advertisedRequest := _node.advertiser.Requests[i][j]
			if advertisedRequest.AssigneeID == "" {
				continue
			}

			assignee := _node.peers[advertisedRequest.AssigneeID]
			if assignee.State.Requests[i][j] || assignee.VirtualState.Requests[i][j] {
				_node.advertiser.Requests[i][j].UUID = ""
				_node.advertiser.Requests[i][j].AssigneeID = ""
			}
		}
	}
	return _node.advertiser
}

func getAssignee(buttonEvent elevio.ButtonEvent, _node node) string {
	if buttonEvent.Button == elevio.BT_Cab {
		return nodeID
	}

	entries := make([]ElevatorEntry, 0)

	entries = append(entries, ElevatorEntry{State: _node.state, Id: nodeID})
	for _, _peer := range _node.peers {
		if !_peer.connected {
			continue
		}

		entries = append(entries, ElevatorEntry{State: _peer.State, Id: _peer.Id})
	}

	return GetBestElevator(entries, buttonEvent)
}

func newAdvertisedRequest(assigneeID string) AdvertisedRequest {
	uuid := uuid.NewString()
	fmt.Printf("Advertising request with UUID %s, assignee %s", uuid, assigneeID)
	return AdvertisedRequest{
		AssigneeID: assigneeID,
		UUID:       uuid,
	}
}
