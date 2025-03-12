package p2p

import (
	"fmt"
	"math/rand"
	elevalgo "sanntidslab/elev_al_go"
	"sanntidslab/p2p/requests"
	"strconv"
	"time"

	"github.com/angrycompany16/driver-go/elevio"
)

const (
	RequestBroadCastPort = 12345
)

// NOTE:
// Read this if the buffer size warning appears
// https://github.com/quic-go/quic-go/wiki/UDP-Buffer-Sizes
// TL;DR
// Run
// sudo sysctl -w net.core.rmem_max=7500000
// and
// sudo sysctl -w net.core.wmem_max=7500000

// TODO: remove ThisNode...etc pattern
// TODO: Implement logging when in single elevator mode
// Requires:
// - Disconn check
// - Activity logging

// TODO: rename utils.go

// NOTE: all fields must be public in structs that are being sent over the network

// TODO: Write getters

// TODO: Make debugging listeners / Network-go easier
// Should be possible to customize a little
// Add requests as its own struct

// What is left to do to get a fully functional elevator system
// 1. Implement the single elevator mode for disconnects
// 2. Do further testing/tuning with packet loss and disconnect
// 3. Test on the lab
// 4. Restructure and clean up the code
// 5. Implement the request assigner correctly

type node struct {
	id               string
	state            *elevalgo.Elevator
	pendingRequests  []requests.PendingRequest
	peerRequests     []requests.PeerRequest
	peers            map[string]peer
	localRequestChan chan requests.RequestInfo // Sent into main.go
}

func (n *node) CheckLostPeers() {
	for _, peer := range n.peers {
		if peer.info.LastSeen.Add(timeout).Before(time.Now()) && peer.info.Connected {
			fmt.Println("Lost peer:", peer)
			peer.info.Connected = false
		}
	}
}

func (n *node) CheckRequests() {
	for i, req := range n.pendingRequests {
		acks := 0
		for _, peer := range n.peers {
			if req.Acks[peer.info.Id] {
				acks++
			}
		}

		if acks == len(n.peers) {
			fmt.Println("Request has been backed up by all other peers")
			n.localRequestChan <- req.RequestInfo

			n.pendingRequests[i] = n.pendingRequests[len(n.pendingRequests)-1]
			n.pendingRequests = n.pendingRequests[:len(n.pendingRequests)-1]
		}
	}
}

// Checks whether the request has been acked - Note that we wait for all peers to
// ack before we take the request
func (n *node) UpdatePendingRequests(lifeSignal Heartbeat) {
	for _, req := range n.pendingRequests {
		fmt.Println("Updating pending request")
		if lifeSignal.WorldView[n.id].State.Requests[req.RequestInfo.Floor][req.RequestInfo.ButtonType] {
			req.Acks[lifeSignal.SenderId] = true
		}
	}
}

func (n *node) Assign(req requests.RequestInfo) string {
	if req.ButtonType == elevio.BT_Cab {
		return n.id
	}

	for _, _peer := range n.peers {
		return _peer.info.Id
	}
	return n.id
}

// TODO: Learn more about go's concurrency patterns and reconsider the peersLock idea
func (n *node) SendRequest(req requests.RequestInfo, assigneeID string) {
	if assigneeID == n.id {
		fmt.Println("Taking the request myself")
		if requests.RequestAlreadyExists(requests.ExtractPendingRequestInfo(n.pendingRequests), req) {
			return
		}
		n.pendingRequests = append(n.pendingRequests, requests.NewPendingRequest(req))
	} else {
		if requests.RequestAlreadyExists(requests.ExtractPeerRequestInfo(n.peerRequests), req) {
			return
		}
		fmt.Println("Appending a request for peer", req)
		n.peerRequests = append(n.peerRequests, requests.NewPeerRequest(req, assigneeID))
	}
}

// TODO: Shorten the names a bit maybe...
func BroadcastPeerRequests(peerRequestChan chan requests.PeerRequest, peerRequests []requests.PeerRequest, peerList map[string]peer) {
	for _, peerRequest := range peerRequests {
		_, ok := peerList[peerRequest.AssigneeID]
		if !ok {
			fmt.Println("Assignee is gone")
			return
		}
		peerRequestChan <- peerRequest
	}
}

func (n *node) TryReceiveRequest(peerRequest requests.PeerRequest) {
	if peerRequest.AssigneeID != n.id {
		return
	}
	n.SelfAssignRequest(peerRequest.RequestInfo)
}

// TODO: Rewrite with map so that duplicate requests are impossible
func (n *node) SelfAssignRequest(request requests.RequestInfo) {
	if requests.RequestAlreadyExists(requests.ExtractPendingRequestInfo(n.pendingRequests), request) {
		return
	}

	n.pendingRequests = append(n.pendingRequests, requests.NewPendingRequest(request))
}

// TODO: Request assigner algorithm is acting kinda sus

func (n *node) GetPeerRequests() []requests.PeerRequest {
	return n.peerRequests
}

func (n *node) GetPeers() map[string]peer {
	return n.peers
}

func InitNode(state *elevalgo.Elevator, id string) node {
	if id == "" {
		r := rand.Int()
		fmt.Println("No id was given. Using randomly generated number", r)
		id = strconv.Itoa(r)
	}

	nodeInstance := newElevator(id, state)

	fmt.Println("Successfully created new network node: ")
	fmt.Println(nodeInstance)

	return nodeInstance
}

func newElevator(id string, state *elevalgo.Elevator) node {
	return node{
		id:              id,
		state:           state,
		peers:           make(map[string]peer, 0),
		peerRequests:    make([]requests.PeerRequest, 0),
		pendingRequests: make([]requests.PendingRequest, 0),
	}
}
