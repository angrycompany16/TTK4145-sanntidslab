package p2p

import (
	"fmt"
	elevalgo "sanntidslab/elev_al_go"
	"sanntidslab/p2p/requests"
	"sanntidslab/utils"
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

// Resolving the cyclic counter-y problems:
// - Note that the problem actually occurs very rarely, as we consider every node
//   to be a single source of truth for its own cab and hall calls
// - The *only* case where a node is not assumed to be correct is if it has crashed,
//   then we need to assume that it has cab calls which it doesn't know, but should
//   be informed about from the other nodes on the network
// - This means that we need to detect when a node has crashed so we can know that the
//   node should have its cab calls overwritten, rather than simply broadcasting "I
//   have zero cab calls"
// - To do this, introduce uptime. Then every backed up request is tagged with the
//   uptime of the node when it was implemented
// - If a node disconnects, its timer will have increased, and so when another node
//   attempts to return the cab calls it will notice that the counter of the node
//   is higher than that of the request, and therefore discard it
// - The node will then broadcast these, and since the uptime value will be larger,
//   the node will overwrite (Take the UNION!!!) with its view of the other node(s)
// - If the node crashes, however, it will return with a lower lifetime than it had
//   before. Then the node will be informed about its lost cab requests, and take these
// - Then these calls are accepted, the node will start to to broadcast it, and then
//   the backups will update the timestamp to be the current timestamp of the node,
//   so when the node is done it willk be considered new information and thus it will
//   overwrite the backed up requests.

// A problematic case?
// - Node A dies, and node B and C are left on the network
// - Node B dies and comes back -> (Node B has no backup of A?)
// - Node C dies and comes back -> (Node C has no backup of A?)

// To resolve this, we have the functionality:
// - If someone has a more recent view backup of the node than we do, we update our
//   backup
// - This way everyone must have the most recent view of the node (under reasonable
//   assumptions), so the case is solved.

// Regarding the arbiter/who should redistribute issue:
// - When a peer is lost, the node looks at the updated peer list
// - If there is a peer with an ID with a lower ID, do nothing
// - Else, redistribute the hall calls that have been backed up

type Heartbeat struct {
	SenderId        string
	State           elevalgo.Elevator
	pendingRequests [elevalgo.NumFloors][elevalgo.NumButtons]requests.PendingRequest // Ack list
	WorldView       map[string]peer                                                  // Our worldview
}

type node struct {
	id              string
	state           elevalgo.Elevator
	pendingRequests [elevalgo.NumFloors][elevalgo.NumButtons]requests.PendingRequest // Ack list
	peerRequests    [elevalgo.NumFloors][elevalgo.NumButtons]int                     // Assignee ID
	peers           map[string]peer
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

	id string,
) {
	// Cases
	// 1. React to life signal
	// 2. Handle a request
	// 3. Send orders to the elevator
	// 4. Send a life signal
	// When? Always.
	// Q: How do we ensure that the node has the right elevator state?
	// A: Iguess... just have it as an input channel?

	// Let's get rid of the request slices

	nodeInstance := InitNode(elevalgo.GetState(), id)

	for {
		select {
		case heartbeat := <-heartbeatChan:
			// TODO: Handle heartbeat
			peers := UpdatePeerList(heartbeat, nodeInstance.peers, id)
			nodeInstance.peers = peers

			// UpdatePendingRequests(heartbeat)
			// nodeInstance.CheckRequests(executeRequestChan)
			// orderChan<-something
		case peerRequest := <-peerRequestChan:
			utils.UNUSED(peerRequest)
			// TODO
			// nodeInstance.TryReceiveRequest(peerRequest)
		case buttonEvent := <-buttonEventChan:
			// TODO: Implement
			// requestInfo := requests.NewRequestInfo(buttonEvent)
			assigneeID := Assign(buttonEvent, nodeInstance)
			nodeInstance.SendRequest(requestInfo, assigneeID)
		case elevatorState := <-elevatorStateChan:
			utils.UNUSED(elevatorState)
		default:
			heartbeat := NewHeartbeat(nodeInstance)
			heartbeatChan <- heartbeat
			peers := CheckLostPeers(nodeInstance.peers)
			nodeInstance.peers = peers

			// p2p.BroadcastPeerRequests(peerRequestChan, nodeInstance.GetPeerRequests(), nodeInstance.GetPeers())
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

func UpdatePeerList(heartbeat Heartbeat, peers map[string]peer, id string) map[string]peer {
	newPeerList := utils.DuplicateMap(peers)

	if id == heartbeat.SenderId {
		return newPeerList
	}

	_peer, ok := newPeerList[heartbeat.SenderId]
	if ok {
		if !_peer.Connected {
			fmt.Println("Reconnecting peer", id)
		}

		_peer.LastSeen = time.Now()
		_peer.State = heartbeat.State

		for i := range elevalgo.NumFloors {
			for j := range elevalgo.NumButtons {
				if !heartbeat.pendingRequests[i][j].Active {
					continue
				}

				if !_peer.State.Requests[i][j] {
					fmt.Println("IT seems someone is having a pending request")
				}
				_peer.State.Requests[i][j] = true
			}
		}

		_peer.Connected = true

		newPeerList[heartbeat.SenderId] = _peer
		return newPeerList
	}

	fmt.Println(heartbeat.SenderId)
	newPeer := newPeer(heartbeat.State, heartbeat.SenderId)
	fmt.Println("New peer created: ")
	fmt.Println(newPeer)

	newPeerList[heartbeat.SenderId] = newPeer

	return newPeerList
}

func CheckLostPeers(peers map[string]peer) map[string]peer {
	newPeerList := utils.DuplicateMap(peers)
	var lostPeer peer
	for _, peer := range newPeerList {
		if peer.LastSeen.Add(timeout).Before(time.Now()) && peer.Connected {
			lostPeer = peer
			lostPeer.Connected = false
			fmt.Println("Lost peer:", peer)
		}
	}

	newPeerList[lostPeer.Id] = lostPeer
	return newPeerList
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
// func UpdatePendingRequests(heartbeat Heartbeat) []requests.PendingRequest {
// 	for _, req := range n.pendingRequests {
// 		if heartbeat.WorldView[n.id].State.Requests[req.RequestInfo.Floor][req.RequestInfo.ButtonType] {
// 			fmt.Println("Updating pending request")
// 			req.Acks[heartbeat.SenderId] = true
// 		}
// 	}
// }

func Assign(buttonEvent elevio.ButtonEvent, self node) string {
	if buttonEvent.Button == elevio.BT_Cab {
		return self.id
	}

	for _, _peer := range self.peers {
		return _peer.Id
	}
	return self.id
}

func (n *node) DistributeRequest(req requests.RequestInfo, assigneeID int) {
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

// func (n *node) TryReceiveRequest(peerRequest requests.PeerRequest) {
// 	if peerRequest.AssigneeID != n.id {
// 		return
// 	}
// 	n.SelfAssignRequest(peerRequest.RequestInfo)
// }

// TODO: Rewrite with map so that duplicate requests are impossible
// func (n *node) SelfAssignRequest(request requests.RequestInfo) {
// 	if requests.RequestAlreadyExists(requests.ExtractPendingRequestInfo(n.pendingRequests), request) {
// 		return
// 	}

// 	n.pendingRequests = append(n.pendingRequests, requests.NewPendingRequest(request))
// }

// TODO: Request assigner algorithm is acting kinda sus

// func (n *node) GetPeerRequests() []requests.PeerRequest {
// 	return n.peerRequests
// }

func (n *node) GetPeers() map[string]peer {
	return n.peers
}

func InitNode(state elevalgo.Elevator, id string) node {
	nodeInstance := node{
		id:    id,
		state: state,
		peers: make(map[string]peer, 0),
	}

	fmt.Println("Successfully created node ", id)

	return nodeInstance
}
