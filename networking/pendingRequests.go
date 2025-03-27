package networking

import (
	"fmt"
	elevalgo "sanntidslab/elevalgo"
	"sanntidslab/elevio"
	"sanntidslab/mapfunctions"
)

// Contains a list of request which are awaiting ack.
type PendingRequestList struct {
	L [elevalgo.NumFloors][elevalgo.NumButtons]PendingRequest
}

type PendingRequest struct {
	acks   map[string]bool // Map of which nodes have acked the request
	Active bool            // Is there a request?
	UUID   string          // UUID given by advertiser, used for avoiding duplicate requests
}

// Accepts a pending requests if all nodes on the network have acknowledged it
func takeAckedRequests(_node node) (elevio.ButtonEvent, PendingRequestList, bool) {
	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			if !_node.pendingRequestList.L[i][j].Active {
				continue
			}

			newAckMap := mapfunctions.DuplicateMap(_node.pendingRequestList.L[i][j].acks)

			if fullyAcked(_node.pendingRequestList.L[i][j], _node.peers) {
				_node.pendingRequestList.L[i][j].Active = false
				_node.pendingRequestList.L[i][j].acks = clearAcks(newAckMap)

				fmt.Printf("Taking request in floor %d, buttontype %s\n", i, elevalgo.ButtonToString(elevio.ButtonType(j)))

				return elevio.ButtonEvent{
						Floor:  i,
						Button: elevio.ButtonType(j),
					},
					_node.pendingRequestList, true
			}
		}
	}
	return elevio.ButtonEvent{
			Floor:  0,
			Button: elevio.ButtonType(0),
		},
		_node.pendingRequestList, false
}

// Updates acknowledgements for pending requests based on heartbeat
func updatePendingRequests(heartbeat Heartbeat, _node node) PendingRequestList {
	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			if !heartbeat.WorldView[nodeID].VirtualState.Requests[i][j] ||
				!_node.pendingRequestList.L[i][j].Active {
				continue
			}

			newAckMap := mapfunctions.DuplicateMap(_node.pendingRequestList.L[i][j].acks)

			if newAckMap[heartbeat.SenderId] {
				continue
			}

			newAckMap[heartbeat.SenderId] = true
			_node.pendingRequestList.L[i][j].acks = newAckMap

			fmt.Printf(" ~ Received ack from node %s ~\n", heartbeat.SenderId)
			printRequest(i, elevio.ButtonType(j))
			fmt.Printf("Current state: %d/%d acks\n\n", countAcks(_node.pendingRequestList.L[i][j]), countConnectedPeers(_node.peers))
		}
	}
	return _node.pendingRequestList
}

func fullyAcked(pendingRequest PendingRequest, peers map[string]peer) bool {
	for _, _peer := range peers {
		if !_peer.connected {
			continue
		}

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

func makePendingRequestList() PendingRequestList {
	var list [elevalgo.NumFloors][elevalgo.NumButtons]PendingRequest

	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			list[i][j] = makePendingRequest()
		}
	}

	return PendingRequestList{L: list}
}

func makePendingRequest() PendingRequest {
	return PendingRequest{
		acks:   make(map[string]bool),
		Active: false,
	}
}

func printRequest(floor int, buttonType elevio.ButtonType) {
	fmt.Printf("Request at floor: %d, button type: %s\n\n", floor, elevalgo.ButtonToString(buttonType))
}
