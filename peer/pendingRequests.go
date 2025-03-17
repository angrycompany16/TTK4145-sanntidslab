package peer

import (
	"fmt"
	elevalgo "sanntidslab/elev_al_go"
	"sanntidslab/mapfunctions"

	"github.com/angrycompany16/driver-go/elevio"
)

type PendingRequest struct {
	acks   map[string]bool
	Active bool
}

type PendingRequests struct {
	List [elevalgo.NumFloors][elevalgo.NumButtons]PendingRequest
}

func takeAckedRequests(_node node) (elevio.ButtonEvent, PendingRequests, bool) {
	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			if !_node.pendingRequests.List[i][j].Active {
				continue
			}

			newAckMap := mapfunctions.DuplicateMap(_node.pendingRequests.List[i][j].acks)

			if fullyAcked(_node.pendingRequests.List[i][j], _node.peers) {
				_node.pendingRequests.List[i][j].Active = false
				_node.pendingRequests.List[i][j].acks = clearAcks(newAckMap)

				fmt.Printf("Taking request in floor %d, buttontype %s\n", i, elevalgo.ButtonToString(elevio.ButtonType(j)))

				return elevio.ButtonEvent{
						Floor:  i,
						Button: elevio.ButtonType(j),
					},
					_node.pendingRequests, true
			}
		}
	}
	return elevio.ButtonEvent{
			Floor:  0,
			Button: elevio.ButtonType(0),
		},
		_node.pendingRequests, false
}

// Updates acks for pending requests based on heartbeat
func updatePendingRequests(heartbeat Heartbeat, _node node) PendingRequests {
	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			if !heartbeat.WorldView[globalID].State.Requests[i][j] {
				continue
			}

			if !_node.pendingRequests.List[i][j].Active {
				continue
			}

			newAckMap := mapfunctions.DuplicateMap(_node.pendingRequests.List[i][j].acks)

			if newAckMap[heartbeat.SenderId] {
				continue
			}

			newAckMap[heartbeat.SenderId] = true
			_node.pendingRequests.List[i][j].acks = newAckMap

			fmt.Printf(" ~ Received ack from node %s ~\n", heartbeat.SenderId)
			PrintRequest(i, elevio.ButtonType(j))
			fmt.Printf("Current state: %d/%d acks\n\n", countAcks(_node.pendingRequests.List[i][j]), countConnectedPeers(_node.peers))
		}
	}
	return _node.pendingRequests
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
	clearedAcks := mapfunctions.DuplicateMap(acks)
	for id := range clearedAcks {
		clearedAcks[id] = false
	}
	return clearedAcks
}

func countAcks(pendingRequest PendingRequest) (sum int) {
	if !pendingRequest.Active {
		return
	}
	for _, ack := range pendingRequest.acks {
		if ack {
			sum++
		}
	}
	return
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

func PrintRequest(floor int, buttonType elevio.ButtonType) {
	fmt.Printf("Request at floor: %d, button type: %s\n", floor, elevalgo.ButtonToString(buttonType))
}
