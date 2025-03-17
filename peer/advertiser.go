package peer

import (
	elevalgo "sanntidslab/elev_al_go"

	"github.com/angrycompany16/driver-go/elevio"
)

type Advertiser struct {
	Requests [elevalgo.NumFloors][elevalgo.NumButtons]string // Contains an ID if a
	// request is being advertised. "" if there is no request being advertised
}

// Stops advertising if a peer is taking the request
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

// TODO: implement
func assign(buttonEvent elevio.ButtonEvent, peers map[string]peer) string {
	if buttonEvent.Button == elevio.BT_Cab {
		return globalID
	}

	for _, _peer := range peers {
		if !_peer.connected {
			continue
		}

		return _peer.Id
	}
	return globalID
}
