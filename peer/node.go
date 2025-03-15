package peer

import (
	"fmt"
	elevalgo "sanntidslab/elev_al_go"
	"time"

	"github.com/angrycompany16/driver-go/elevio"
)

var (
	timeout  = time.Millisecond * 500
	GlobalID string /* NOTE: This variable is global because only written to on init */
)

// NOTE: Scary bug: Sometimes it seems that the peers connect and disconnect constantly,
// but i have no idea how to reproduce the bug???

// TODO: Implement logging when in single elevator mode
// Requires:
// - Disconn check
// - Activity logging

// TODO: rename utils.go

// TODO: Kinda the last piece of the puzzle: Implement this
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
//   so when the node is done it will be considered new information and thus it will
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
// - If there is a peer with an ID with a lower uptime, do nothing
// - Else, redistribute the hall calls that have been backed up

// NOTE: all fields must be public in structs that are being sent over the network

type Heartbeat struct {
	SenderId        string
	State           elevalgo.Elevator
	Uptime          int64
	WorldView       map[string]peer
	PendingRequests PendingRequests
}

type node struct {
	state           elevalgo.Elevator
	pendingRequests PendingRequests
	advertiser      Advertiser
	peers           map[string]peer
	uptime          int64
}

// TODO: Function is now 50+ lines of complex code... probably want so subdivide this
func NodeProcess(
	advertiserChan chan Advertiser,
	heartbeatChan chan Heartbeat,
	buttonEventChan <-chan elevio.ButtonEvent,
	elevatorStateChan <-chan elevalgo.Elevator,
	orderChan chan<- elevio.ButtonEvent,

	initState elevalgo.Elevator,
) {
	nodeInstance := newNode(initState)

	for {
		select {
		case heartbeat := <-heartbeatChan:
			updatedPeers, newPeer := updatePeerList(heartbeat, nodeInstance.peers)
			nodeInstance.peers = updatedPeers

			updatedAdvertiser := updateAdvertiser(nodeInstance.peers, nodeInstance.advertiser)

			updatedPendingRequests := updatePendingRequests(heartbeat, nodeInstance.state, nodeInstance.pendingRequests)

			if newPeer {
				updatedPendingRequests = restoreLostCabCalls(updatedPendingRequests, heartbeat, nodeInstance.uptime)
			}

			order, clearedPendingRequests, ok := takeAckedRequests(updatedPendingRequests, nodeInstance.peers)

			nodeInstance.pendingRequests = clearedPendingRequests
			nodeInstance.advertiser = updatedAdvertiser

			// No need to send requests that haven't been acked only to ignore them later
			if !ok {
				continue
			}
			orderChan <- order
		case advertiser := <-advertiserChan:
			newPendingRequestList := tryReceiveRequest(advertiser, nodeInstance.pendingRequests, nodeInstance.state)

			nodeInstance.pendingRequests = newPendingRequestList
		case buttonEvent := <-buttonEventChan:
			assigneeID := assign(buttonEvent, nodeInstance.peers)

			newPendingRequests, newAdvertiser := distributeRequest(buttonEvent, assigneeID, nodeInstance.pendingRequests, nodeInstance.advertiser)

			nodeInstance.advertiser = newAdvertiser
			nodeInstance.pendingRequests = newPendingRequests
		case elevatorState := <-elevatorStateChan:
			// Update pending requests
			nodeInstance.state = elevatorState
		default:
			heartbeat := newHeartbeat(nodeInstance)
			heartbeatChan <- heartbeat
			nodeInstance.uptime++

			updatedPeers, lostPeer := checkLostPeers(nodeInstance.peers)
			nodeInstance.peers = updatedPeers

			// If a peer is lost, redistribute backed up requests
			newPendingRequests, newAdvertiser := redistributeLostRequests(lostPeer, updatedPeers, nodeInstance.pendingRequests, nodeInstance.advertiser, nodeInstance.uptime)
			nodeInstance.pendingRequests = newPendingRequests
			nodeInstance.advertiser = newAdvertiser

			advertiserChan <- nodeInstance.advertiser
		}
	}
}

func newHeartbeat(node node) Heartbeat {
	return Heartbeat{
		SenderId:        GlobalID,
		State:           node.state,
		Uptime:          node.uptime,
		PendingRequests: node.pendingRequests,
		WorldView:       node.peers,
	}
}

// TODO: This is not very well written
func distributeRequest(
	buttonEvent elevio.ButtonEvent,
	assigneeID string,
	pendingRequests PendingRequests,
	advertiser Advertiser,
) (
	PendingRequests,
	Advertiser,
) {
	if assigneeID == GlobalID {
		fmt.Println("Taking the request myself")
		pendingRequests.List[buttonEvent.Floor][buttonEvent.Button] = 0
		return pendingRequests, advertiser
	} else {
		fmt.Println("Sending request to", assigneeID)
		advertiser.Requests[buttonEvent.Floor][buttonEvent.Button] = assigneeID
		return pendingRequests, advertiser
	}
}

func redistributeLostRequests(lostPeer peer, peers map[string]peer, pendingRequests PendingRequests, advertiser Advertiser, uptime int64) (PendingRequests, Advertiser) {
	// TODO: Implement safety check so that both alive nodes don't redistribute the
	// lost requests. This is to avoid duplication
	newPendingRequests := pendingRequests
	newAdvertiser := advertiser

	for _, _peer := range peers {
		if !_peer.connected {
			continue
		}

		// If we are not the oldest connected peer, do nothing
		if _peer.Uptime > uptime {
			return pendingRequests, advertiser
		}
	}

	for i := range elevalgo.NumFloors {
		// TODO: Make generic wrt. number of cab buttons
		for j := range elevalgo.NumButtons - 1 {
			if lostPeer.State.Requests[i][j] {
				buttonEvent := elevio.ButtonEvent{i, elevio.ButtonType(j)}
				assigneeID := assign(buttonEvent, peers)
				newPendingRequests, newAdvertiser = distributeRequest(buttonEvent, assigneeID, newPendingRequests, newAdvertiser)
			}
		}
	}

	return newPendingRequests, newAdvertiser
}

func restoreLostCabCalls(pendingRequests PendingRequests, heartbeat Heartbeat, uptime int64) PendingRequests {
	if heartbeat.SenderId == GlobalID || heartbeat.Uptime < uptime {
		return pendingRequests
	}

	for i := range elevalgo.NumFloors {
		// TODO: Make generic number of cab buttons
		if heartbeat.WorldView[GlobalID].State.Requests[i][elevalgo.NumButtons-1] {
			pendingRequests.List[i][elevalgo.NumButtons-1] = 0 // Set the cab button
			// to have an active request
		}
	}
	return pendingRequests
}

func tryReceiveRequest(advertiser Advertiser, pendingRequests PendingRequests, state elevalgo.Elevator) PendingRequests {
	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			if advertiser.Requests[i][j] != GlobalID {
				continue
			}

			if state.Requests[i][j] {
				continue
			}

			pendingRequests.List[i][j] = 0
		}
	}
	return pendingRequests
}

func countConnectedPeers(peers map[string]peer) (connectedPeers int) {
	for _, _peer := range peers {
		if _peer.connected {
			connectedPeers++
		}
	}
	return connectedPeers
}

func newNode(state elevalgo.Elevator) node {
	nodeInstance := node{
		state:           state,
		peers:           make(map[string]peer, 0),
		pendingRequests: makePendingRequestList(),
		uptime:          0,
	}

	fmt.Println("Successfully created node ", GlobalID)

	return nodeInstance
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

func makePendingRequestList() PendingRequests {
	return PendingRequests{
		List: [elevalgo.NumFloors][elevalgo.NumButtons]int{
			{-1, -1, -1},
			{-1, -1, -1},
			{-1, -1, -1},
			{-1, -1, -1},
		},
	}
}
