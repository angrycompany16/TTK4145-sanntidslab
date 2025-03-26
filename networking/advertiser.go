package networking

import (
	"fmt"
	elevalgo "sanntidslab/elevalgo"
	"sanntidslab/elevio"

	"github.com/google/uuid"
)

// Contains a list of requests being advertised to other peers
type Advertiser struct {
	Requests [elevalgo.NumFloors][elevalgo.NumButtons]AdvertisedRequest
}

type AdvertisedRequest struct {
	AssigneeID string
	UUID       string
}

// Stops advertising if a peer is actively taking the request
func updateAdvertiser(_node node) Advertiser {
	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			advertisedRequest := _node.advertiser.Requests[i][j]
			if advertisedRequest.AssigneeID == "" {
				continue
			}

			// TODO: Somehow shorten this down a bit maybe
			if _node.peers[advertisedRequest.AssigneeID].State.Requests[i][j] {
				fmt.Println("Stop advertising")
				_node.advertiser.Requests[i][j].UUID = ""
				_node.advertiser.Requests[i][j].AssigneeID = ""
			} else if _node.peers[advertisedRequest.AssigneeID].VirtualState.Requests[i][j] {
				fmt.Println("Stop advertising")
				_node.advertiser.Requests[i][j].UUID = ""
				_node.advertiser.Requests[i][j].AssigneeID = ""
			}
		}
	}
	return _node.advertiser
}

func assignRequest(buttonEvent elevio.ButtonEvent, _node node) string {
	if buttonEvent.Button == elevio.BT_Cab {
		return nodeID
	}

	entries := make([]elevalgo.ElevatorEntry, 0)

	entries = append(entries, elevalgo.ElevatorEntry{State: _node.state, Id: nodeID})
	for _, _peer := range _node.peers {
		if !_peer.connected {
			continue
		}

		entries = append(entries, elevalgo.ElevatorEntry{State: _peer.State, Id: _peer.Id})
	}

	return elevalgo.GetBestElevator(entries, buttonEvent)
}

func newAdvertisedRequest(assigneeID string) AdvertisedRequest {
	uuid := uuid.NewString()
	fmt.Printf("Advertising request with UUID %s, assignee %s", uuid, assigneeID)
	return AdvertisedRequest{
		AssigneeID: assigneeID,
		UUID:       uuid,
	}
}
