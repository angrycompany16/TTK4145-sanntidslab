package p2p

import (
	"fmt"
	elevalgo "sanntidslab/elev_al_go"
	"sanntidslab/utils"
	"time"
)

type peer struct {
	State     elevalgo.Elevator
	Id        int
	LastSeen  time.Time
	Connected bool
}

// TODO: Might be easier to not use Connected
func newPeer(state elevalgo.Elevator, id int) peer {
	return peer{
		State:     state,
		Id:        id,
		LastSeen:  time.Now(),
		Connected: false,
	}
}

func (p peer) String() string {
	return fmt.Sprintf("------- Peer ----\n ~ id: %s\n", p.Id)
}

func (n *node) ExtractPeerState() map[int]elevalgo.Elevator {
	return utils.MapMap(
		n.peers,
		func(_peer peer) elevalgo.Elevator { return _peer.State },
	)
}
