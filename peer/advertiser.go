package peer

import (
	"fmt"
	elevalgo "sanntidslab/elev_al_go"

	"github.com/angrycompany16/driver-go/elevio"
)

type Advertiser struct {
	Requests [elevalgo.NumFloors][elevalgo.NumButtons]string // Contains an ID if a
	// request is being advertised. "" if there is no request being advertised
}

func updateAdvertiser(peers map[string]peer, advertiser Advertiser) Advertiser {
	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			assigneeID := advertiser.Requests[i][j]
			if assigneeID == "" {
				continue
			}

			if peers[assigneeID].State.Requests[i][j] {
				fmt.Println("Removing advertised request")
				advertiser.Requests[i][j] = ""
			}
		}
	}
	return advertiser
}

// For now this is very simple, self assign if cab and otherwise assign to first
// connected peer
func assign(buttonEvent elevio.ButtonEvent, peers map[string]peer) string {
	if buttonEvent.Button == elevio.BT_Cab {
		return GlobalID
	}

	for _, _peer := range peers {
		if !_peer.connected {
			continue
		}

		return _peer.Id
	}
	return GlobalID
}
