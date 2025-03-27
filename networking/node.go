package networking

import (
	"fmt"
	elevalgo "sanntidslab/elevalgo"
	"sanntidslab/elevio"
	"sanntidslab/network-driver/bcast"
	"time"
)

const (
	heartbeatBroadcastPort = 36251 // For peer detection and updating
	requestBroadCastPort   = 12345 // For advertising requests
)

// Declared outside the node struct for convenience
var (
	nodeID string
	uptime int64
)

// Contains the information needed to distribute, receive and ack messages over the
// network
type node struct {
	state              elevalgo.Elevator
	pendingRequestList PendingRequestList
	advertiser         Advertiser
	peers              map[string]peer
}

// Runs a networking node. Distributes & acknowledges messages while maintaining a list
// of peers on the network
func RunNode(
	buttonEventChan <-chan elevio.ButtonEvent,
	elevatorStateChan <-chan elevalgo.Elevator,
	orderChan chan<- elevio.ButtonEvent,
	peerStates chan<- []elevalgo.Elevator,

	elevatorConfig elevalgo.Config,
	id string,
) {
	nodeID = id
	uptime = 0
	nodeInstance := newNode(elevatorConfig)

	advertiserChan := make(chan Advertiser)
	heartbeatTx := make(chan Heartbeat, 1)
	heartbeatRx := make(chan Heartbeat, 1)

	go bcast.Transmitter(heartbeatBroadcastPort, heartbeatTx)
	go bcast.Receiver(heartbeatBroadcastPort, heartbeatRx)

	go bcast.Transmitter(requestBroadCastPort, advertiserChan)
	go bcast.Receiver(requestBroadCastPort, advertiserChan)

	for {
		select {
		case heartbeat := <-heartbeatRx:
			var peerIsAdded bool
			nodeInstance.peers, peerIsAdded = checkNewPeers(heartbeat, nodeInstance.peers)

			var peerIsUpdated bool
			nodeInstance.peers, peerIsUpdated = updateExistingPeers(heartbeat, nodeInstance.peers)

			nodeInstance.advertiser = updateAdvertiser(nodeInstance)
			nodeInstance.pendingRequestList = updatePendingRequests(heartbeat, nodeInstance)

			if peerIsAdded {
				nodeInstance.pendingRequestList = restoreLostCabCalls(heartbeat, nodeInstance)
			}

			if peerIsUpdated {
				peerStates <- extractPeerStates(nodeInstance.peers)
			}
		case advertiser := <-advertiserChan:
			nodeInstance.pendingRequestList = takeAdvertisedCalls(advertiser, nodeInstance)
		case buttonEvent := <-buttonEventChan:
			assigneeID := getAssignee(buttonEvent, nodeInstance)

			nodeInstance = distributeRequest(buttonEvent, assigneeID, nodeInstance)
		case elevatorState := <-elevatorStateChan:
			nodeInstance.state = elevatorState
		default:
			heartbeat := newHeartbeat(nodeInstance)
			heartbeatTx <- heartbeat
			uptime++

			var lostPeer peer
			var hasLostPeer bool
			nodeInstance.peers, lostPeer, hasLostPeer = checkLostPeers(nodeInstance.peers)

			if hasLostPeer {
				peerStates <- extractPeerStates(nodeInstance.peers)
				nodeInstance = redistributeLostHallCalls(lostPeer, nodeInstance)
			}

			order, clearedPendingRequests, hasOrder := takeAckedRequests(nodeInstance)
			nodeInstance.pendingRequestList = clearedPendingRequests

			advertiserChan <- nodeInstance.advertiser

			if hasOrder {
				orderChan <- order
			}
			time.Sleep(time.Millisecond * 10)
		}
	}
}

func distributeRequest(buttonEvent elevio.ButtonEvent, assigneeID string, _node node) node {
	if assigneeID == nodeID {
		fmt.Println("Self-assigned request:")
		printRequest(buttonEvent.Floor, buttonEvent.Button)
		_node.pendingRequestList.L[buttonEvent.Floor][buttonEvent.Button].Active = true
		return _node
	} else {
		printRequest(buttonEvent.Floor, buttonEvent.Button)
		_node.advertiser.Requests[buttonEvent.Floor][buttonEvent.Button] = newAdvertisedRequest(assigneeID)
		return _node
	}
}

func redistributeLostHallCalls(lostPeer peer, _node node) node {
	// If we are not the oldest connected peer, we do nothing in order to avoid
	// duplicating calls
	for _, _peer := range _node.peers {
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
				buttonEvent := elevio.ButtonEvent{Floor: i, Button: elevio.ButtonType(j)}
				assigneeID := getAssignee(buttonEvent, _node)
				_node = distributeRequest(buttonEvent, assigneeID, _node)
			}
		}
	}
	return _node
}

// Restores lost cab calls from heartbeat if the information is not outdated
func restoreLostCabCalls(heartbeat Heartbeat, _node node) PendingRequestList {
	// Ignore cab call backups if the information is outdated (e.g. in case of
	// disconnect)
	if heartbeat.SenderId == nodeID || heartbeat.Uptime < uptime {
		return _node.pendingRequestList
	}

	fmt.Println("Restoring cab calls")

	for i := range elevalgo.NumFloors {
		if heartbeat.WorldView[nodeID].BackedUpCabCalls[i] {
			_node.pendingRequestList.L[i][2].Active = true

			fmt.Println("Received lost cab call from", heartbeat.SenderId)
			printRequest(i, elevio.BT_Cab)
		}
	}
	return _node.pendingRequestList
}

// Accepts a call being advertised by another node
func takeAdvertisedCalls(otherAdvertiser Advertiser, _node node) PendingRequestList {
	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			if otherAdvertiser.Requests[i][j].AssigneeID != nodeID ||
				_node.state.Requests[i][j] ||
				_node.pendingRequestList.L[i][j].UUID == otherAdvertiser.Requests[i][j].UUID {
				continue
			}

			fmt.Println("Taking advertised request, ID:", otherAdvertiser.Requests[i][j].UUID)
			printRequest(i, elevio.ButtonType(j))
			_node.pendingRequestList.L[i][j].Active = true
			_node.pendingRequestList.L[i][j].UUID = otherAdvertiser.Requests[i][j].UUID
		}
	}
	return _node.pendingRequestList
}

func newNode(elevatorConfig elevalgo.Config) node {
	nodeInstance := node{
		state:              elevalgo.NewUninitializedElevator(elevatorConfig),
		peers:              make(map[string]peer, 0),
		pendingRequestList: makePendingRequestList(),
	}

	fmt.Println("Successfully created node ", nodeID)

	return nodeInstance
}
