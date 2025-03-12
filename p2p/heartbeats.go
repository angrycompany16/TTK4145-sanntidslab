package p2p

import (
	"fmt"
	elevalgo "sanntidslab/elev_al_go"
	"sanntidslab/p2p/requests"
	"time"
)

type Heartbeat struct {
	SenderId        string
	State           elevalgo.Elevator
	pendingRequests []requests.PendingRequest
	WorldView       map[string]PeerInfo // Our worldview
}

func (n *node) SendHeartbeat(heartbeatChan chan Heartbeat) {
	heartbeat := Heartbeat{
		SenderId:        n.id,
		State:           *n.state,
		pendingRequests: n.pendingRequests,
		WorldView:       n.ExtractPeerInfo(),
	}

	heartbeatChan <- heartbeat
}

func (n *node) ReceiveHeartbeat(heartbeat Heartbeat) {
	if n.id == heartbeat.SenderId {
		return
	}

	_peer, ok := n.peers[heartbeat.SenderId]
	if ok {
		_peer.info.LastSeen = time.Now()
		_peer.info.State = heartbeat.State

		// Backs up the unserviced requests from other elevators
		for _, req := range heartbeat.pendingRequests {
			_peer.info.State.Requests[req.RequestInfo.Floor][req.RequestInfo.ButtonType] = true
		}

		n.UpdatePendingRequests(heartbeat)
		n.CheckRequests()

		return
	}

	newPeer := newPeer(heartbeat.State, heartbeat.SenderId)

	n.peers[heartbeat.SenderId] = newPeer
	fmt.Println("New peer added: ")
	fmt.Println(newPeer)

	// n.GetLostRequests(heartbeat)
}
