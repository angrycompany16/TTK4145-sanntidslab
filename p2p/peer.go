package p2p

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
	Connected bool
}

// TODO: Might be easier to not use Connected
func newPeer(state elevalgo.Elevator, id string) peer {
	return peer{
		State:     state,
		Id:        id,
		LastSeen:  time.Now(),
		Connected: true,
	}
}

func (p peer) String() string {
	return fmt.Sprintf("------- Peer ----\n ~ id: %s\n", p.Id)
}

func (n *node) ExtractPeerState() map[string]elevalgo.Elevator {
	return utils.MapMap(
		n.peers,
		func(_peer peer) elevalgo.Elevator { return _peer.State },
	)
}
