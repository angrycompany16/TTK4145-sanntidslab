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

// TODO: rename utils.go

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

// Problem:
// We have two competing opinions on the state of the node
// If there is a pending request, we say that the node's state should be updated
// However, the node should also be considered a single source of truth for its own calls, and this
// will disagree on pending requests as they haven't been accepted yet
// The solution (this is actually not a problem!)
// We first update the peer with what it broadcasts its own state to be, and then if it also broadcasted
// some pending requests we add those in after.

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
			fmt.Println("Received button press")
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
		pendingRequests.List[buttonEvent.Floor][buttonEvent.Button].Active = true
		return pendingRequests, advertiser
	} else {
		fmt.Println("Sending request to", assigneeID)
		advertiser.Requests[buttonEvent.Floor][buttonEvent.Button] = assigneeID
		return pendingRequests, advertiser
	}
}

func redistributeLostRequests(lostPeer peer, peers map[string]peer, pendingRequests PendingRequests, advertiser Advertiser, uptime int64) (PendingRequests, Advertiser) {
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
			pendingRequests.List[i][elevalgo.NumButtons-1].Active = true // Set the cab button
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

			if pendingRequests.List[i][j].Active {
				continue
			}

			pendingRequests.List[i][j].Active = true
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
		pendingRequests: makePendingRequests(),
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
