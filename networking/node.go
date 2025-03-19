package networking

import (
	"fmt"
	elevalgo "sanntidslab/elev_al_go"

	"github.com/angrycompany16/Network-go/network/transfer"
	"github.com/angrycompany16/driver-go/elevio"
)

const (
	stateBroadcastPort   = 36251 // Akkordrekke
	requestBroadCastPort = 12345 // Just a random number
)

var (
	nodeID string
	uptime int64
)

// NOTE: Scary bug: Sometimes it seems that the peers connect and disconnect constantly,
// but i have no idea how to reproduce the bug???

// NOTE: all fields must be public in structs that are being sent over the network

// Contains the information needed to distribute, receive and ack messages over the
// network
type node struct {
	state              elevalgo.Elevator
	pendingRequestList PendingRequestList
	advertiser         Advertiser
	peers              map[string]peer
}

// TODO: Function is now 50+ lines of complex code... probably want so subdivide this
// Runs a networking node. Distributes & acknowledges messages while maintaining a list
// of peers on the network
func RunNode(
	buttonEventChan <-chan elevio.ButtonEvent,
	elevatorStateChan <-chan elevalgo.Elevator,
	orderChan chan<- elevio.ButtonEvent,
	peerStates chan<- []elevalgo.Elevator,

	initState elevalgo.Elevator,
	id string,
) {
	nodeInstance := newNode(initState)
	nodeID = id
	uptime = 0

	advertiserChan := make(chan Advertiser)
	heartbeatChan := make(chan Heartbeat, 1) // Buffered to avoid bugs where nodes
	// disconnect when the system is very busy

	go transfer.BroadcastSender(stateBroadcastPort, heartbeatChan)
	go transfer.BroadcastReceiver(stateBroadcastPort, heartbeatChan)

	go transfer.BroadcastSender(requestBroadCastPort, advertiserChan)
	go transfer.BroadcastReceiver(requestBroadCastPort, advertiserChan)

	for {
		select {
		case heartbeat := <-heartbeatChan:
			var addedPeer bool
			nodeInstance.peers, addedPeer = checkNewPeers(heartbeat, nodeInstance.peers)
			nodeInstance.peers = updateExistingPeers(heartbeat, nodeInstance.peers)

			nodeInstance.advertiser = updateAdvertiser(nodeInstance)
			nodeInstance.pendingRequestList = updatePendingRequests(heartbeat, nodeInstance)

			if addedPeer {
				nodeInstance.pendingRequestList = restoreLostCabCalls(heartbeat, nodeInstance)
			}
		case advertiser := <-advertiserChan:
			nodeInstance.pendingRequestList = takeAdvertisedCalls(advertiser, nodeInstance)
		case buttonEvent := <-buttonEventChan:
			assigneeID := assignRequest(buttonEvent, nodeInstance)

			nodeInstance = distributeRequest(buttonEvent, assigneeID, nodeInstance)
		case elevatorState := <-elevatorStateChan:
			nodeInstance.state = elevatorState
		default:
			heartbeat := newHeartbeat(nodeInstance)
			heartbeatChan <- heartbeat
			uptime++

			var lostPeer peer
			nodeInstance.peers, lostPeer = checkLostPeers(nodeInstance.peers)

			nodeInstance = redistributeLostHallCalls(lostPeer, nodeInstance)

			order, clearedPendingRequests, hasOrder := takeAckedRequests(nodeInstance)
			nodeInstance.pendingRequestList = clearedPendingRequests

			advertiserChan <- nodeInstance.advertiser

			if hasOrder {
				fmt.Println("Giving order")
				orderChan <- order
			}

			peerStates <- ExtractPeerStates(nodeInstance.peers)
		}
	}
}

func distributeRequest(buttonEvent elevio.ButtonEvent, assigneeID string, _node node) node {
	if assigneeID == nodeID {
		fmt.Println("Self-assigned request:")
		printRequest(buttonEvent.Floor, buttonEvent.Button)
		fmt.Println()
		_node.pendingRequestList.L[buttonEvent.Floor][buttonEvent.Button].Active = true
		return _node
	} else {
		fmt.Println("Sending request to peer", assigneeID)
		printRequest(buttonEvent.Floor, buttonEvent.Button)
		fmt.Println()
		_node.advertiser.Requests[buttonEvent.Floor][buttonEvent.Button] = assigneeID
		return _node
	}
}

// Problem: When we reconnect, we'll first discover one node, get our calls restored
// and then look through our list of peers (which has only one node) and get one
// single ack and thus take it. This means that other nodes don't ack, but is
// this a small enough edge case for it to not matter?
// No this is fine actually, because right after we get the same request, but this time
// with two peers, so it's going to get acked by both peers anyway

// An actual problem: Consider the following sequence:
// - Node A gets a cab call
// - Node A dies
// - Node B gets a cab call
// - Node B dies
// - Node A comes back, completes backed up cab call
// - Node C then dies (containing B's backup)
// - Node B comes back. It has then lost the request

// Possible method for resolving
// When updating our view of a node, if the node is not connected we accept that
// the node with the most recent update becomes the single source of truth
// In other words we need to make update peer function a bit more complicated

// The problem is fucking RESOLVED

func redistributeLostHallCalls(lostPeer peer, _node node) node {
	if (lostPeer == peer{}) {
		return _node
	}

	for _, _peer := range _node.peers {
		// If we are not the oldest connected peer, we do nothing to avoid duplicating
		// calls
		if !_peer.connected {
			continue
		}

		if _peer.Uptime > uptime {
			return _node
		}
	}

	fmt.Println("Redistributing hall calls from peer", lostPeer.Id)

	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons - elevalgo.NumCabButtons {
			if lostPeer.State.Requests[i][j] {
				buttonEvent := elevio.ButtonEvent{i, elevio.ButtonType(j)}
				assigneeID := assignRequest(buttonEvent, _node)
				_node = distributeRequest(buttonEvent, assigneeID, _node)
			}
		}
	}

	return _node
}

func restoreLostCabCalls(heartbeat Heartbeat, _node node) PendingRequestList {
	if heartbeat.SenderId == nodeID || heartbeat.Uptime < uptime {
		return _node.pendingRequestList
	}

	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumCabButtons {
			if heartbeat.WorldView[nodeID].State.Requests[i][elevalgo.NumButtons-j-1] {
				_node.pendingRequestList.L[i][elevalgo.NumButtons-j-1].Active = true

				fmt.Println("Received lost cab call from", heartbeat.SenderId)
				printRequest(i, elevio.BT_Cab)
			}
		}
	}
	return _node.pendingRequestList
}

func takeAdvertisedCalls(otherAdvertiser Advertiser, _node node) PendingRequestList {
	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			if otherAdvertiser.Requests[i][j] != nodeID ||
				_node.state.Requests[i][j] ||
				_node.pendingRequestList.L[i][j].Active {
				continue
			}

			_node.pendingRequestList.L[i][j].Active = true
		}
	}
	return _node.pendingRequestList
}

func newNode(state elevalgo.Elevator) node {
	nodeInstance := node{
		state:              state,
		peers:              make(map[string]peer, 0),
		pendingRequestList: makePendingRequestList(),
	}

	fmt.Println("Successfully created node ", nodeID)

	return nodeInstance
}
