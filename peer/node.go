package peer

import (
	"fmt"
	elevalgo "sanntidslab/elev_al_go"

	"github.com/angrycompany16/Network-go/network/transfer"
	"github.com/angrycompany16/driver-go/elevio"
)

const (
	stateBroadcastPort   = 36251 // Akkordrekke
	requestBroadCastPort = 12345
)

var (
	globalID string
)

// NOTE: Scary bug: Sometimes it seems that the peers connect and disconnect constantly,
// but i have no idea how to reproduce the bug???

// TODO: rename utils.go

// NOTE: all fields must be public in structs that are being sent over the network
type node struct {
	state           elevalgo.Elevator
	pendingRequests PendingRequests
	advertiser      Advertiser
	peers           map[string]peer
	uptime          int64
}

// TODO: Function is now 50+ lines of complex code... probably want so subdivide this
func NodeProcess(
	buttonEventChan <-chan elevio.ButtonEvent,
	elevatorStateChan <-chan elevalgo.Elevator,
	orderChan chan<- elevio.ButtonEvent,

	initState elevalgo.Elevator,
	id string,
) {
	nodeInstance := newNode(initState)
	globalID = id

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
			nodeInstance.peers, addedPeer = updatePeerList(heartbeat, nodeInstance.peers)

			nodeInstance.advertiser = updateAdvertiser(nodeInstance)
			nodeInstance.pendingRequests = updatePendingRequests(heartbeat, nodeInstance)

			if addedPeer {
				nodeInstance.pendingRequests = restoreLostCabCalls(heartbeat, nodeInstance)
			}
		case advertiser := <-advertiserChan:
			nodeInstance.pendingRequests = takeAdvertisedCalls(advertiser, nodeInstance)
		case buttonEvent := <-buttonEventChan:
			assigneeID := assign(buttonEvent, nodeInstance.peers)

			nodeInstance = distributeRequest(buttonEvent, assigneeID, nodeInstance)
		case elevatorState := <-elevatorStateChan:
			nodeInstance.state = elevatorState
		default:
			heartbeatChan <- newHeartbeat(nodeInstance)
			nodeInstance.uptime++

			var lostPeer peer
			nodeInstance.peers, lostPeer = checkLostPeers(nodeInstance.peers)

			nodeInstance = redistributeRequests(lostPeer, nodeInstance)

			order, clearedPendingRequests, hasOrder := takeAckedRequests(nodeInstance)
			nodeInstance.pendingRequests = clearedPendingRequests

			advertiserChan <- nodeInstance.advertiser

			if hasOrder {
				orderChan <- order
			}
		}
	}
}

func newHeartbeat(node node) Heartbeat {
	return Heartbeat{
		SenderId:        globalID,
		State:           node.state,
		Uptime:          node.uptime,
		PendingRequests: node.pendingRequests,
		WorldView:       node.peers,
	}
}

func distributeRequest(
	buttonEvent elevio.ButtonEvent,
	assigneeID string,
	_node node,
) node {
	if assigneeID == globalID {
		fmt.Println("Self-assigned request:")
		PrintRequest(buttonEvent.Floor, buttonEvent.Button)
		fmt.Println()
		_node.pendingRequests.List[buttonEvent.Floor][buttonEvent.Button].Active = true
		return _node
	} else {
		fmt.Println("Sending request to peer", assigneeID)
		PrintRequest(buttonEvent.Floor, buttonEvent.Button)
		fmt.Println()
		_node.advertiser.Requests[buttonEvent.Floor][buttonEvent.Button] = assigneeID
		return _node
	}
}

func redistributeRequests(lostPeer peer, _node node) node {
	for _, _peer := range _node.peers {
		if !_peer.connected {
			continue
		}

		// If we are not the oldest connected peer, do nothing
		if _peer.Uptime > _node.uptime {
			return _node
		}
	}

	for i := range elevalgo.NumFloors {
		// TODO: Make generic wrt. number of cab buttons
		for j := range elevalgo.NumButtons - 1 {
			if lostPeer.State.Requests[i][j] {
				buttonEvent := elevio.ButtonEvent{i, elevio.ButtonType(j)}
				assigneeID := assign(buttonEvent, _node.peers)
				_node = distributeRequest(buttonEvent, assigneeID, _node)
			}
		}
	}

	return _node
}

func restoreLostCabCalls(heartbeat Heartbeat, _node node) PendingRequests {
	if heartbeat.SenderId == globalID || heartbeat.Uptime < _node.uptime {
		return _node.pendingRequests
	}

	for i := range elevalgo.NumFloors {
		// TODO: Make generic number of cab buttons
		if heartbeat.WorldView[globalID].State.Requests[i][elevalgo.NumButtons-1] {
			_node.pendingRequests.List[i][elevalgo.NumButtons-1].Active = true // Set the cab button
			// to have an active request
		}
	}
	return _node.pendingRequests
}

func takeAdvertisedCalls(otherAdvertiser Advertiser, _node node) PendingRequests {
	for i := range elevalgo.NumFloors {
		for j := range elevalgo.NumButtons {
			if otherAdvertiser.Requests[i][j] != globalID {
				continue
			}

			if _node.state.Requests[i][j] {
				continue
			}

			if _node.pendingRequests.List[i][j].Active {
				continue
			}

			_node.pendingRequests.List[i][j].Active = true
		}
	}
	return _node.pendingRequests
}

func newNode(state elevalgo.Elevator) node {
	nodeInstance := node{
		state:           state,
		peers:           make(map[string]peer, 0),
		pendingRequests: makePendingRequests(),
		uptime:          0,
	}

	fmt.Println("Successfully created node ", globalID)

	return nodeInstance
}
