package p2p

import (
	"fmt"
	elevalgo "sanntidslab/elev_al_go"
	"sanntidslab/p2p/requests"
	"time"

	"github.com/angrycompany16/driver-go/elevio"
)

const (
	RequestBroadCastPort = 12345
)

var (
	timeout = time.Millisecond * 500
)

// TODO: Implement logging when in single elevator mode
// Requires:
// - Disconn check
// - Activity logging

// TODO: rename utils.go

// NOTE: all fields must be public in structs that are being sent over the network

// TODO: Write getters

type Heartbeat struct {
	SenderId        int
	State           elevalgo.Elevator
	pendingRequests []requests.PendingRequest
	WorldView       map[int]peer // Our worldview
}

type node struct {
	id              int
	state           elevalgo.Elevator
	pendingRequests []requests.PendingRequest
	peerRequests    []requests.PeerRequest // Requests to ID with floor/button
	peers           map[int]peer
}

// TODO: Probably split
func NodeProcess(
	/* I/O channels */
	heartbeatChan chan Heartbeat,
	peerRequestChan chan requests.PeerRequest,
	/* input only channels */
	buttonEventChan <-chan elevio.ButtonEvent,
	elevatorStateChan <-chan elevalgo.Elevator,
	/* output only channels */
	orderChan chan<- requests.RequestInfo,

	id int,
) {
	// Cases
	// 1. React to life signal
	// 2. Handle a request
	// 3. Send orders to the elevator
	// 4. Send a life signal
	// When? Always.
	// Q: How do we ensure that the node has the right elevator state?
	// A: Iguess... just have it as an input channel?

	nodeInstance := InitNode(elevalgo.GetState(), id)

	for {
		select {
		case heartbeat := <-heartbeatChan:
			// TODO: Handle heartbeat
			peers := UpdatePeerList(heartbeat, nodeInstance.peers, id)
			nodeInstance.peers = peers

			UpdatePendingRequests(heartbeat)
			// nodeInstance.CheckRequests(executeRequestChan)
			// orderChan<-something
		case peerRequest := <-peerRequestChan:
			// TODO
			// nodeInstance.TryReceiveRequest(peerRequest)
		case buttonEvent := <-buttonEventChan:
			// TODO: Implement
			// requestInfo := requests.NewRequestInfo(buttonEvent)
			// assigneeID := nodeInstance.Assign(requestInfo)
			// nodeInstance.SendRequest(requestInfo, assigneeID)
		case elevatorState := <-elevatorStateChan:

		default:
			heartbeat := NewHeartbeat(nodeInstance)
			heartbeatChan <- heartbeat

			// nodeInstance.CheckLostPeers()
			// p2p.BroadcastPeerRequests(peerRequestChan, nodeInstance.GetPeerRequests(), nodeInstance.GetPeers())
			// timer.CheckTimeout()
		}
	}
}

func NewHeartbeat(node node) Heartbeat {
	return Heartbeat{
		SenderId:        node.id,
		State:           node.state,
		pendingRequests: node.pendingRequests,
		WorldView:       node.peers,
	}
}

func UpdatePeerList(heartbeat Heartbeat, peers map[int]peer, id int) map[int]peer {
	// If me, do nothing
	if id == heartbeat.SenderId {
		return peers
	}

	_peer, ok := peers[heartbeat.SenderId]
	if ok {
		_peer.LastSeen = time.Now()
		_peer.State = heartbeat.State

		// Loop through requests
		for _, req := range heartbeat.pendingRequests {
			_peer.State.Requests[req.RequestInfo.Floor][req.RequestInfo.ButtonType] = true
		}

		return peers
	}

	newPeer := newPeer(heartbeat.State, heartbeat.SenderId)
	fmt.Println("New peer created: ")
	fmt.Println(newPeer)
	peers[heartbeat.SenderId] = newPeer
	return peers
}

func CheckLostPeers(peers map[int]peer) map[int]peer {
	for _, peer := range peers {
		if peer.LastSeen.Add(timeout).Before(time.Now()) && peer.Connected {
			fmt.Println("Lost peer:", peer)
			peer.Connected = false
		}
	}
}

func (n *node) CheckRequests(executeRequestChan chan requests.RequestInfo) {
	// for i, req := range n.pendingRequests {
	// 	acks := 0
	// for _, peer := range n.peers {
	// if req.Acks[peer.info.Id] {
	// 	acks++
	// }
	// }

	// if acks == len(n.peers) {
	// 	fmt.Println("Request has been backed up by all other peers")
	// 	executeRequestChan <- req.RequestInfo

	// 	// TODO: Can be rewritten
	// 	n.pendingRequests[i] = n.pendingRequests[len(n.pendingRequests)-1]
	// 	n.pendingRequests = n.pendingRequests[:len(n.pendingRequests)-1]
	// }
	// }
}

// Checks whether the request has been acked - Note that we wait for all peers to
// ack before we take the request
func UpdatePendingRequests(heartbeat Heartbeat) []requests.PendingRequest {
	for _, req := range n.pendingRequests {
		if heartbeat.WorldView[n.id].State.Requests[req.RequestInfo.Floor][req.RequestInfo.ButtonType] {
			fmt.Println("Updating pending request")
			req.Acks[heartbeat.SenderId] = true
		}
	}
}

func (n *node) Assign(req requests.RequestInfo) int {
	if req.ButtonType == elevio.BT_Cab {
		return n.id
	}

	for _, _peer := range n.peers {
		return _peer.info.Id
	}
	return n.id
}

// TODO: Learn more about go's concurrency patterns and reconsider the peersLock idea
func (n *node) SendRequest(req requests.RequestInfo, assigneeID int) {
	// if assigneeID == n.id {
	// 	fmt.Println("Taking the request myself")
	// 	if requests.RequestAlreadyExists(requests.ExtractPendingRequestInfo(n.pendingRequests), req) {
	// 		return
	// 	}
	// 	n.pendingRequests = append(n.pendingRequests, requests.NewPendingRequest(req))
	// } else {
	// 	if requests.RequestAlreadyExists(requests.ExtractPeerRequestInfo(n.peerRequests), req) {
	// 		return
	// 	}
	// 	fmt.Println("Appending a request for peer", req)
	// 	n.peerRequests = append(n.peerRequests, requests.NewPeerRequest(req, assigneeID))
	// }
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

func (n *node) GetPeers() map[int]peer {
	return n.peers
}

func InitNode(state elevalgo.Elevator, id int) node {
	nodeInstance := node{
		id:    id,
		state: state,
		peers: make(map[int]peer, 0),
	}

	fmt.Println("Successfully created node ", id)

	return nodeInstance
}
