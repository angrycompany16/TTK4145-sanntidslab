package networking

import (
	"fmt"
	elevalgo "sanntidslab/elevalgo"
	"sanntidslab/mapfunctions"
	"time"
)

var (
	timeout = time.Millisecond * 500
)

type peer struct {
	Id           string
	Uptime       int64
	LastSeen     time.Time
	connected    bool
	State        elevalgo.Elevator
	VirtualState elevalgo.Elevator // Contains "desired state" which is only used
	// when acking requests
	BackedUpCabCalls [elevalgo.NumFloors]bool
}

// Adds or restores peers from received heartbeat
func checkNewPeers(heartbeat Heartbeat, peers map[string]peer) (map[string]peer, bool) {
	newPeerList := mapfunctions.DuplicateMap(peers)
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
		fmt.Println("Restored peer from worldview: peer", id)
	}

	if hasRestoredPeer {
		newPeerList[restoredPeer.Id] = restoredPeer
	}

	return newPeerList, true
}

// Updates peer list with info from heartbeat
func updateExistingPeers(heartbeat Heartbeat, peers map[string]peer) (newPeerList map[string]peer, updated bool) {
	newPeerList = mapfunctions.DuplicateMap(peers)
	updated = false

	if nodeID == heartbeat.SenderId {
		return
	}

	updatedPeer, ok := newPeerList[heartbeat.SenderId]
	if !ok {
		return
	}

	if !updatedPeer.connected {
		fmt.Println("Reconnecting pear", updatedPeer.Id)
	}

	if heartbeat.State.Requests != updatedPeer.State.Requests {
		updated = true
		updatedPeer.State = heartbeat.State
	} else {
		updatedPeer.State = heartbeat.State
	}

	updatedPeer.LastSeen = time.Now()
	updatedPeer.Uptime = heartbeat.Uptime
	updatedPeer.connected = true

	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			updatedPeer.VirtualState.Requests[i][j] = heartbeat.PendingRequests.L[i][j].Active
		}
		// If the peer is actively looking for backup, we no longer need to back it up in the BackedUpCabCalls
		// Instead its backed up in the peer itself
		if heartbeat.PendingRequests.L[i][2].Active {
			updatedPeer.BackedUpCabCalls[i] = false
		}
	}

	newPeerList[heartbeat.SenderId] = updatedPeer
	return
}

func checkLostPeers(peers map[string]peer) (newPeerList map[string]peer, lostPeer peer, hasLostPeer bool) {
	newPeerList = mapfunctions.DuplicateMap(peers)
	hasLostPeer = false

	for _, peer := range newPeerList {
		if peer.LastSeen.Add(timeout).Before(time.Now()) && peer.connected {
			lostPeer = peer
			lostPeer.connected = false
			hasLostPeer = true
			fmt.Println("Lost peer", peer.Id)
		}
	}

	if hasLostPeer {
		lostPeer.BackedUpCabCalls = elevalgo.ExtractCabCalls(lostPeer.State)
		newPeerList[lostPeer.Id] = lostPeer
	}
	return
}

func countConnectedPeers(peers map[string]peer) (connectedPeers int) {
	for _, _peer := range peers {
		if _peer.connected {
			connectedPeers++
		}
	}
	return connectedPeers
}

func extractPeerStates(peers map[string]peer) (states []elevalgo.Elevator) {
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
