package networking

import (
	elevalgo "sanntidslab/elev_al_go"

	"github.com/angrycompany16/driver-go/elevio"
)

// Contains a list of requests being advertised to other peers
type Advertiser struct {
	Requests [elevalgo.NumFloors][elevalgo.NumButtons]string
}

// Stops advertising if a peer is actively taking the request
func updateAdvertiser(_node node) Advertiser {
	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			assigneeID := _node.advertiser.Requests[i][j]
			if assigneeID == "" {
				continue
			}

			if _node.peers[assigneeID].State.Requests[i][j] {
				_node.advertiser.Requests[i][j] = ""
			}
		}
	}
	return _node.advertiser
}

func assignRequest(buttonEvent elevio.ButtonEvent, _node node) string {
	if buttonEvent.Button == elevio.BT_Cab {
		return nodeID
	}

	elevators := make([]elevalgo.Elevator, 0)
	ids := make([]string, 0)

	elevators = append(elevators, _node.state)
	ids = append(ids, nodeID)
	for _, _peer := range _node.peers {
		if !_peer.connected {
			continue
		}

		elevators = append(elevators, _peer.State)
		ids = append(ids, _peer.Id)
	}

	orderList := elevalgo.GetBestOrder(elevators)
	return ids[orderList[0]]
}
