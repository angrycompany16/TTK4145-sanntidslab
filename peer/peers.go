package peer

import (
	"fmt"
	elevalgo "sanntidslab/elev_al_go"
	"sanntidslab/mapfunctions"
	"time"
)

var (
	timeout = time.Millisecond * 500
)

type Heartbeat struct {
	SenderId        string
	State           elevalgo.Elevator
	Uptime          int64
	WorldView       map[string]peer
	PendingRequests PendingRequests
}

type peer struct {
	State     elevalgo.Elevator
	Id        string
	LastSeen  time.Time
	Uptime    int64
	connected bool
}

func updatePeerList(heartbeat Heartbeat, peers map[string]peer) (map[string]peer, bool) {
	newPeerList := mapfunctions.DuplicateMap(peers)

	if globalID == heartbeat.SenderId {
		return newPeerList, false
	}

	_peer, ok := newPeerList[heartbeat.SenderId]
	if ok {
		if !_peer.connected {
			fmt.Println("Reconnecting pear", globalID)
		}

		_peer.LastSeen = time.Now()
		_peer.State = heartbeat.State
		_peer.Uptime = heartbeat.Uptime

		for i := range elevalgo.NumFloors {
			for j := range elevalgo.NumButtons {
				if !heartbeat.PendingRequests.List[i][j].Active {
					continue
				}

				_peer.State.Requests[i][j] = true
			}
		}

		_peer.connected = true

		newPeerList[heartbeat.SenderId] = _peer
		return newPeerList, false
	}

	newPeer := newPeer(heartbeat.State, heartbeat.SenderId, heartbeat.Uptime)
	fmt.Println("New peer created: peer", newPeer.Id)

	newPeerList[heartbeat.SenderId] = newPeer

	return newPeerList, true
}

func checkLostPeers(peers map[string]peer) (map[string]peer, peer) {
	newPeerList := mapfunctions.DuplicateMap(peers)
	var lostPeer peer
	hasLostPeer := false
	// TODO: Check if there is a heartbeat waiting
	for _, peer := range newPeerList {
		if peer.LastSeen.Add(timeout).Before(time.Now()) && peer.connected {
			lostPeer = peer
			lostPeer.connected = false
			hasLostPeer = true
			fmt.Println("Lost peer", peer.Id)
		}
	}

	if hasLostPeer {
		newPeerList[lostPeer.Id] = lostPeer
	}
	return newPeerList, lostPeer
}

func countConnectedPeers(peers map[string]peer) (connectedPeers int) {
	for _, _peer := range peers {
		if _peer.connected {
			connectedPeers++
		}
	}
	return connectedPeers
}

func newPeer(state elevalgo.Elevator, id string, uptime int64) peer {
	return peer{
		State:     state,
		Id:        id,
		LastSeen:  time.Now(),
		Uptime:    uptime,
		connected: true,
	}
}
