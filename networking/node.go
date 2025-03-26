package networking

import (
	"fmt"
	elevalgo "sanntidslab/elevalgo"
	"sanntidslab/elevio"
	"sanntidslab/network-driver/bcast"
	"time"
)

const (
	stateBroadcastPort   = 36251 // Akkordrekke
	requestBroadCastPort = 12345 // Just a random number
)

var (
	nodeID string
	uptime int64
)

// NOTE: Another scary bug: Sometimes the system randomly spawns in a lot of
// nonexistnet requests on startup???
// Phantom button presses seem to be fairly rare. They are somehow spawning from
// the driver
// How tf??
// For some reason this bug is very very hard to reproduce

// NOTE: Scary bug: Sometimes it seems that the peers connect and disconnect constantly,
// but i have no idea how to reproduce the bug???

// A problem: If we have *very* high (90%) packet loss on the request broadcast port,
// it's essentially impossible for a hall request to be taken, because no one will
// pick up on the advertiser's requests.
// I'm not sure if this is something that needs to be considered though...

// NOTE: all fields must be public in structs that are being sent over the network

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

	initState elevalgo.Elevator,
	id string,
) {
	nodeInstance := newNode(initState)
	nodeID = id
	uptime = 0

	advertiserChan := make(chan Advertiser)
	heartbeatTx := make(chan Heartbeat, 1)
	heartbeatRx := make(chan Heartbeat, 1)

	go bcast.Transmitter(stateBroadcastPort, heartbeatTx)
	go bcast.Receiver(stateBroadcastPort, heartbeatRx)

	go bcast.Transmitter(requestBroadCastPort, advertiserChan)
	go bcast.Receiver(requestBroadCastPort, advertiserChan)

	for {
		select {
		case heartbeat := <-heartbeatRx:
			var addedPeer bool
			nodeInstance.peers, addedPeer = checkNewPeers(heartbeat, nodeInstance.peers)

			var updatedPeer bool
			nodeInstance.peers, updatedPeer = updateExistingPeers(heartbeat, nodeInstance.peers)

			nodeInstance.advertiser = updateAdvertiser(nodeInstance)
			nodeInstance.pendingRequestList = updatePendingRequests(heartbeat, nodeInstance)
			if addedPeer {
				nodeInstance.pendingRequestList = restoreLostCabCalls(heartbeat, nodeInstance)
			}

			if updatedPeer {
				peerStates <- extractPeerStates(nodeInstance.peers)
			}
		case advertiser := <-advertiserChan:
			nodeInstance.pendingRequestList = takeAdvertisedCalls(advertiser, nodeInstance)
		case buttonEvent := <-buttonEventChan:
			fmt.Println("Button event:", buttonEvent)
			assigneeID := assignRequest(buttonEvent, nodeInstance)

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
				fmt.Println("Giving order")
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
		fmt.Println()
		_node.pendingRequestList.L[buttonEvent.Floor][buttonEvent.Button].Active = true
		return _node
	} else {
		printRequest(buttonEvent.Floor, buttonEvent.Button)
		fmt.Println()
		_node.advertiser.Requests[buttonEvent.Floor][buttonEvent.Button] = newAdvertisedRequest(assigneeID)
		return _node
	}
}

func redistributeLostHallCalls(lostPeer peer, _node node) node {
	// If we are not the oldest connected peer, we do nothing to avoid duplicating
	// calls
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
				assigneeID := assignRequest(buttonEvent, _node)
				_node = distributeRequest(buttonEvent, assigneeID, _node)
			}
		}
	}

	return _node
}

func restoreLostCabCalls(heartbeat Heartbeat, _node node) PendingRequestList {
	// TODO: Maybe check heartbeat.World...uptime instead of Uptime
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

func newNode(state elevalgo.Elevator) node {
	nodeInstance := node{
		state:              state,
		peers:              make(map[string]peer, 0),
		pendingRequestList: makePendingRequestList(),
	}

	fmt.Println("Successfully created node ", nodeID)

	return nodeInstance
}
