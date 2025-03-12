package p2p

import (
	"fmt"
	elevalgo "sanntidslab/elev_al_go"
	"sanntidslab/utils"
	"time"
)

var (
	timeout = time.Millisecond * 500
	// ThisNode node
	// LifeSignalChan = make(chan Heartbeat)
)

type peer struct {
	info PeerInfo
}

type PeerInfo struct {
	State     elevalgo.Elevator
	Id        string
	LastSeen  time.Time
	Connected bool
}

func newPeer(state elevalgo.Elevator, id string) peer {
	return peer{
		info: newPeerInfo(state, id),
	}
}

func newPeerInfo(state elevalgo.Elevator, id string) PeerInfo {
	return PeerInfo{
		State:     state,
		Id:        id,
		LastSeen:  time.Now(),
		Connected: false,
	}
}

func (p peer) String() string {
	return fmt.Sprintf("------- Peer ----\n ~ id: %s\n", p.info.Id)
}

func (n *node) ExtractPeerInfo() map[string]PeerInfo {
	return utils.MapMap(
		n.peers,
		func(_peer peer) PeerInfo { return _peer.info },
	)
}

func (n *node) ExtractPeerState() map[string]elevalgo.Elevator {
	return utils.MapMap(
		n.peers,
		func(_peer peer) elevalgo.Elevator { return _peer.info.State },
	)
}
