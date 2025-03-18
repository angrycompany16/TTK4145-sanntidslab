package networking

import (
	"fmt"
	elevalgo "sanntidslab/elev_al_go"
	"sanntidslab/listfunctions"
	"time"
)

var (
	timeout = time.Millisecond * 500
)

type peer struct {
	State     elevalgo.Elevator
	Id        string
	LastSeen  time.Time
	Uptime    int64
	connected bool
}

// Main problem: If we die, we don't have the data for which peers have previously
// been on the network
// Conclusion: We need to add peers back to the network based on the worldviews of
// alive nodes

// Adds a new peer to the list of peers if it doesn't already exist
func checkNewPeers(heartbeat Heartbeat, peers map[string]peer) (map[string]peer, bool) {
	newPeerList := listfunctions.DuplicateMap(peers)
	_, exists := newPeerList[heartbeat.SenderId]

	if nodeID == heartbeat.SenderId || exists {
		return newPeerList, false
	}

	newPeer := newPeer(heartbeat)
	fmt.Println("New peer created: peer", newPeer.Id)
	newPeerList[heartbeat.SenderId] = newPeer

	hasRestoredPeer := false
	var restoredPeer peer
	for id, _peer := range heartbeat.WorldView {
		_, exists := newPeerList[id]
		if exists || id == nodeID {
			continue
		}

		hasRestoredPeer = true
		restoredPeer = _peer
		fmt.Println("Restored peer from heartbeat: peer", id)
	}

	if hasRestoredPeer {
		newPeerList[restoredPeer.Id] = restoredPeer
	}

	return newPeerList, true
}

// Updates peer list with info from heartbeat
func updateExistingPeers(heartbeat Heartbeat, peers map[string]peer) map[string]peer {
	newPeerList := listfunctions.DuplicateMap(peers)

	if nodeID == heartbeat.SenderId {
		return newPeerList
	}

	// Loop through disconnected ndoes
	// If the heartbeat has older info than we do about it, accept that info
	updatedPeer, ok := newPeerList[heartbeat.SenderId]
	if !ok {
		return newPeerList
	}

	if !updatedPeer.connected {
		fmt.Println("Reconnecting pear", updatedPeer.Id)
	}

	updatedPeer.LastSeen = time.Now()
	updatedPeer.State = heartbeat.State
	updatedPeer.Uptime = heartbeat.Uptime
	updatedPeer.connected = true

	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			if !heartbeat.PendingRequests.L[i][j].Active {
				continue
			}

			updatedPeer.State.Requests[i][j] = true
		}
	}

	newPeerList[heartbeat.SenderId] = updatedPeer
	return newPeerList
}

func checkLostPeers(peers map[string]peer) (newPeerList map[string]peer, lostPeer peer) {
	newPeerList = listfunctions.DuplicateMap(peers)
	hasLostPeer := false

	for _, peer := range newPeerList {
		if peer.LastSeen.Add(timeout).Before(time.Now()) && peer.connected {
			lostPeer = peer
			lostPeer.connected = false
			hasLostPeer = true
			fmt.Println("Lost peer", peer.Id)
		}
	}

	// Note: Go doesn't allow you to modify a map while iterating through it, so updating
	// the peer list has to be done like this
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

func ExtractPeerStates(peers map[string]peer) (states []elevalgo.Elevator) {
	for _, _peer := range peers {
		if _peer.connected {
			states = append(states, _peer.State)
		}
	}
	return
}

func newPeer(heartbeat Heartbeat) peer {
	return peer{
		State:     heartbeat.State,
		Id:        heartbeat.SenderId,
		LastSeen:  time.Now(),
		Uptime:    heartbeat.Uptime,
		connected: true,
	}
}
