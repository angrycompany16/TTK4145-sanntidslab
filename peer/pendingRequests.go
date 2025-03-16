package peer

import (
	"fmt"
	elevalgo "sanntidslab/elev_al_go"
	"sanntidslab/utils"

	"github.com/angrycompany16/driver-go/elevio"
)

type PendingRequest struct {
	acks   map[string]bool
	Active bool
}

type PendingRequests struct {
	// TODO: either check continuously or replace with map[string]bool
	// However we also need some way of representing whether it is inactive...
	List [elevalgo.NumFloors][elevalgo.NumButtons]PendingRequest // -1 represents inactive,
	// otherwise this represents the number of elevators who have backed up this request
}

// BUG: Sometimes the elevator takes requests without ack
// Sometimes it also seems that acks don't arrive

func takeAckedRequests(pendingRequests PendingRequests, peers map[string]peer) (elevio.ButtonEvent, PendingRequests, bool) {
	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			if !pendingRequests.List[i][j].Active {
				continue
			}

			newAckMap := utils.DuplicateMap(pendingRequests.List[i][j].acks)

			if fullyAcked(pendingRequests.List[i][j], peers) {
				pendingRequests.List[i][j].Active = false
				pendingRequests.List[i][j].acks = clearAcks(newAckMap)

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

			// TODO: Maybe remove the check for whether we have already taken the request?
			if !pendingRequests.List[i][j].Active || state.Requests[i][j] {
				continue
			}

			newAckMap := utils.DuplicateMap(pendingRequests.List[i][j].acks)
			// Already acked?
			if newAckMap[heartbeat.SenderId] {
				continue
			}

			// TODO: Visualizer of how many acks are received?
			fmt.Println("Received ack from node", heartbeat.SenderId)
			newAckMap[heartbeat.SenderId] = true
			pendingRequests.List[i][j].acks = newAckMap
		}
	}
	return pendingRequests
}

func fullyAcked(pendingRequest PendingRequest, peers map[string]peer) bool {
	for _, _peer := range peers {
		if !pendingRequest.acks[_peer.Id] {
			return false
		}
	}
	return true
}

func clearAcks(acks map[string]bool) map[string]bool {
	clearedAcks := utils.DuplicateMap(acks)
	for id := range clearedAcks {
		clearedAcks[id] = false
	}
	return clearedAcks
}

func makePendingRequests() PendingRequests {
	var list [elevalgo.NumFloors][elevalgo.NumButtons]PendingRequest

	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			list[i][j] = makePendingRequest()
		}
	}

	return PendingRequests{List: list}
}

func makePendingRequest() PendingRequest {
	return PendingRequest{
		acks:   make(map[string]bool),
		Active: false,
	}
}
