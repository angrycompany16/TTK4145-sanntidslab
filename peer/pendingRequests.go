package peer

import (
	"fmt"
	elevalgo "sanntidslab/elev_al_go"

	"github.com/angrycompany16/driver-go/elevio"
)

type PendingRequests struct {
	List [elevalgo.NumFloors][elevalgo.NumButtons]int // -1 represents inactive,
	// otherwise this represents the number of elevators who have backed up this request
}

func takeAckedRequests(pendingRequests PendingRequests, peers map[string]peer) (elevio.ButtonEvent, PendingRequests, bool) {
	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			if pendingRequests.List[i][j] == -1 {
				continue
			}

			if pendingRequests.List[i][j] == countConnectedPeers(peers) {
				fmt.Println("Taking pending request")
				pendingRequests.List[i][j] = -1
				return elevio.ButtonEvent{
						Floor:  i,
						Button: elevio.ButtonType(j),
					},
					pendingRequests, true
			}
		}
	}
	return elevio.ButtonEvent{
			Floor:  0,
			Button: elevio.ButtonType(0),
		},
		pendingRequests, false
}

func updatePendingRequests(heartbeat Heartbeat, state elevalgo.Elevator, pendingRequests PendingRequests) PendingRequests {
	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			// TODO: heartbeat.WorldView[id].State.Requests[i][j] is hard to read
			if !heartbeat.WorldView[GlobalID].State.Requests[i][j] {
				continue
			}

			if pendingRequests.List[i][j] == -1 || state.Requests[i][j] {
				continue
			}
			fmt.Println("Increased acks")
			pendingRequests.List[i][j]++ // This request has been backed up
		}
	}
	return pendingRequests
}
