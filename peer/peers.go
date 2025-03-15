package peer

import (
	"fmt"
	elevalgo "sanntidslab/elev_al_go"
	"sanntidslab/utils"
	"time"
)

type peer struct {
	State     elevalgo.Elevator
	Id        string
	LastSeen  time.Time
	Uptime    int64
	connected bool
}

// TODO: For now we just return an empty peer on update; maybe this could be done better?
// Could take inspo from the PeerUpdate thingy
func updatePeerList(heartbeat Heartbeat, peers map[string]peer) (map[string]peer, bool) {
	newPeerList := utils.DuplicateMap(peers)

	if GlobalID == heartbeat.SenderId {
		return newPeerList, false
	}

	_peer, ok := newPeerList[heartbeat.SenderId]
	if ok {
		if !_peer.connected {
			fmt.Println("Reconnecting pear", GlobalID)
		}

		_peer.LastSeen = time.Now()
		_peer.State = heartbeat.State
		_peer.Uptime = heartbeat.Uptime

		for i := range elevalgo.NumFloors {
			for j := range elevalgo.NumButtons {
				if heartbeat.PendingRequests.List[i][j] == -1 {
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
	fmt.Println("New peer created: ")
	fmt.Println(newPeer)

	newPeerList[heartbeat.SenderId] = newPeer

	return newPeerList, true
}

func checkLostPeers(peers map[string]peer) (map[string]peer, peer) {
	newPeerList := utils.DuplicateMap(peers)
	var lostPeer peer
	hasLostPeer := false
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
