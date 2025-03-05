package distribute

import (
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"reflect"
	elevalgo "sanntidslab/elev_al_go"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/angrycompany16/Network-go/network/localip"
	"github.com/angrycompany16/Network-go/network/transfer"
	"github.com/angrycompany16/driver-go/elevio"
	"github.com/mitchellh/mapstructure"
)

// TODO: Maybe switch to peer / backup system

// NOTE:
// Read this if the buffer size warning appears
// https://github.com/quic-go/quic-go/wiki/UDP-Buffer-Sizes
// TL;DR
// Run
// sudo sysctl -w net.core.rmem_max=7500000
// and
// sudo sysctl -w net.core.wmem_max=7500000

// TODO: remove This... pattern

const (
	stateBroadcastPort = 36251 // Akkordrekke
)

var (
	timeout        = time.Second * 5
	ThisNode       node
	LifeSignalChan = make(chan LifeSignal)
)

// NOTE: all fields must be public in structs that are being sent over the network
type LifeSignal struct {
	ListenerAddr       net.UDPAddr
	SenderId           string
	State              elevalgo.Elevator
	UnservicedRequests []ElevatorRequest
	WorldView          []PeerInfo // Our worldview
}

type ElevatorRequest struct {
	SenderId   string            `json:"SenderId"`
	ButtonType elevio.ButtonType `json:"ButtonType"`
	Floor      int               `json:"Floor"`
}

type Ack struct{}

// any message we send should be turned into one of these
// TODO: Consider moving this into Network-go?
type GeneralMsg struct {
	TypeName string      `json:"TypeName"`
	Data     interface{} `json:"Data"`
}

// TODO: Make debugging listeners / Network-go easier
// Should be possible to customize a little
// Add requests as its own struct
type node struct {
	id              string
	state           *elevalgo.Elevator
	ip              net.IP
	requestListener transfer.Listener
	peers           []*peer
	peersLock       *sync.Mutex
	// BackupAckChan    chan Ack
	localRequestChan   chan ElevatorRequest
	unservicedRequests []ElevatorRequest
	// localState       elevalgo.Elevator // The state the elevator should be in
}

// Problem: I want the request sender in the peer struct, but it would be nice to instantiate
// a backup sender / listener at the same time
// Solution: Maybe separate the connection parts from each other so that we can ...
// That's going to be difficult, because the sending/connecting is tightly coupled
// with the peer struct
// Perhaps we can move the peer somewhere else and use only the node structs?
// Not sure, it is quite a strange conundrum
// For now, let's keep them in the same file
type PeerInfo struct {
	State     elevalgo.Elevator
	Id        string
	LastSeen  time.Time
	Connected bool
}

type peer struct {
	sender transfer.Sender
	info   PeerInfo
}

func (n *node) timeout() {
	for {
		n.peersLock.Lock()
		for _, peer := range n.peers {
			if peer.info.LastSeen.Add(timeout).Before(time.Now()) && peer.info.Connected {
				// Better idea: Set peer node state to disconnected
				// This allows us to reconnect easily
				peer.info.Connected = false
				// TODO: Try removing this to see if it's actually unnecessary
				fmt.Println("Disabling send channel to peer:", peer)
				peer.sender.QuitChan <- 1
				n.requestListener.QuitChan <- peer.info.Id
				// n.peers[i] = n.peers[len(n.peers)-1]
				// n.peers = n.peers[:len(n.peers)-1]
			}
		}
		n.peersLock.Unlock()
	}
}

func (n *node) sendLifeSignal(signalChan chan (LifeSignal)) {
	for {
		// TODO: move this out into its own function
		derefPeers := make([]PeerInfo, 0)
		for _, peer := range n.peers {

			// for i := 0; i < elevalgo.NumFloors; i++ {
			// 	for j := 0; j < elevalgo.NumButtons; j++ {
			// 		if peer.info.State.Requests[i][j] {
			// 			fmt.Println("Something is backed up")
			// 		}
			// 	}
			// }

			derefPeers = append(derefPeers, peer.info)
		}
		signal := LifeSignal{
			ListenerAddr:       n.requestListener.Addr,
			SenderId:           n.id,
			State:              *n.state,
			UnservicedRequests: n.unservicedRequests,
			WorldView:          derefPeers,
		}

		signalChan <- signal
		time.Sleep(time.Millisecond * 10)
	}
}

// We need to somehow read the requests from the information given by the other peers:
// My idea: Just take me OR given info
func (n *node) readLifeSignals(signalChan chan (LifeSignal)) {
LifeSignals:
	for lifeSignal := range signalChan {
		if n.id == lifeSignal.SenderId {
			continue
		}

		n.peersLock.Lock()
		for _, _peer := range n.peers {
			if _peer.info.Id == lifeSignal.SenderId {
				if !_peer.info.Connected {
					fmt.Println("Connect peer")
					n.ConnectPeer(_peer, lifeSignal)
					n.GetLostRequests(lifeSignal)
					n.peersLock.Unlock()
					continue LifeSignals
				}

				_peer.info.LastSeen = time.Now()
				_peer.info.State = lifeSignal.State

				// TODO: [IN network.go] we can safely remove sender.Connnected
				// If we receive an unserviced request, update the worldview
				for _, req := range lifeSignal.UnservicedRequests {
					fmt.Println("Someone else had an unserviced request")
					_peer.info.State.Requests[req.Floor][req.ButtonType] = true
				}

				// Check if our unserviced requests have been backed up
				n.CheckUnservicedRequests(lifeSignal)

				n.peersLock.Unlock()

				continue LifeSignals
			}
		}

		requestSender := transfer.NewSender(lifeSignal.ListenerAddr, n.id)

		newPeer := newPeer(requestSender, lifeSignal.State, lifeSignal.SenderId)

		n.peers = append(n.peers, newPeer)
		fmt.Println("New peer added: ")
		fmt.Println(newPeer)

		n.peersLock.Unlock()
	}
}

func (n *node) ConnectPeer(_peer *peer, lifeSignal LifeSignal) {
	_peer.sender.Addr = lifeSignal.ListenerAddr
	go _peer.sender.Send()
	<-_peer.sender.ReadyChan
	_peer.info.Connected = true
}

func (n *node) GetLostRequests(lifeSignal LifeSignal) {
	for _, _peer := range lifeSignal.WorldView {
		if _peer.Id != n.id {
			continue
		}

		for i := 0; i < elevalgo.NumFloors; i++ {
			for j := 0; j < elevalgo.NumButtons; j++ {
				// fmt.Println(_peer.State.Requests[i][j])
				if !_peer.State.Requests[i][j] {
					continue
				}
				n.localRequestChan <- ElevatorRequest{
					SenderId:   "",
					ButtonType: elevio.ButtonType(j),
					Floor:      i,
				}
			}
		}
	}
}

func (n *node) CheckUnservicedRequests(lifeSignal LifeSignal) {
	for _, _peer := range lifeSignal.WorldView {
		// This is the sender's view of me
		if _peer.Id == n.id {
			for i, req := range n.unservicedRequests {
				if _peer.State.Requests[req.Floor][req.ButtonType] {
					// The request has been backed up
					// Time to service the request
					fmt.Println("Someone has backed up my unserviced request, I can now service it!")
					n.localRequestChan <- req
					// NOTE: This will reorder the requests
					n.unservicedRequests[i] = n.unservicedRequests[len(n.peers)-1]
					n.unservicedRequests = n.unservicedRequests[:len(n.peers)-1]
				}
			}
		}
	}
}

// Other idea: We can give the elevator a list of requests, and then when we compare
// the state we can check if someone is broadcasting the request. If they are, send it
// back to main.go and we have our guarantee.
func (n *node) TakeRequest(button elevio.ButtonEvent) {
	req := ElevatorRequest{
		SenderId:   n.id,
		ButtonType: button.Button,
		Floor:      button.Floor,
	}

	assigneeID := n.getBestElevator()

	if assigneeID == n.id {
		// TODO: check if the request already exists
		if slices.Contains(n.unservicedRequests, req) {
			return
		}
		n.unservicedRequests = append(n.unservicedRequests, req)
	}
}

func (n *node) SelfRequestNode(button elevio.ButtonEvent) ElevatorRequest {
	req := ElevatorRequest{
		SenderId:   n.id,
		ButtonType: button.Button,
		Floor:      button.Floor,
	}

	return req
}

// TODO: This module is getting way too big. It needs to be more fine-grained.
// TODO: Request assigner algorithm is acting kinda sus
// TODO: Can surely be written in a nicer way
// Returns the id of the elevator that should take the request
func (n *node) getBestElevator() string {
	return n.id

	// elevatorList := []elevalgo.Elevator{*n.state}
	// for _, peer := range n.peers {
	// 	elevatorList = append(elevatorList, peer.state)
	// }

	// var winnerID string
	// n.peersLock.Lock() // Changing the peer list during this operation is probably not a good
	// orderList := elevalgo.GetBestOrder(elevatorList)
	// if orderList[0] == 0 {
	// 	winnerID = n.id
	// } else {
	// 	winnerID = n.peers[orderList[0]-1].id
	// }
	// n.peersLock.Unlock()

	// return winnerID
}

// TODO: Can simplify a lot
func (n *node) PipeListener(requestRx chan ElevatorRequest, ackRx chan Ack, recordRx chan Record) {
	for msg := range n.requestListener.DataChan {
		var message GeneralMsg
		mapstructure.Decode(msg, &message)
		switch message.TypeName {
		case reflect.TypeOf(ElevatorRequest{}).Name():
			var request ElevatorRequest
			err := mapstructure.Decode(message.Data, &request)
			if err != nil {
				log.Fatal("Died")
			}
			requestRx <- request
		case reflect.TypeOf(Record{}).Name():
			var record Record
			err := mapstructure.Decode(message.Data, &record)
			if err != nil {
				log.Fatal("Died record??")
			}
			recordRx <- record
		}
	}
}

func InitNode(state *elevalgo.Elevator, localRequestChan chan ElevatorRequest, id string) {
	for {
		if id == "" {
			r := rand.Int()
			fmt.Println("No id was given. Using randomly generated number", r)
			id = strconv.Itoa(r)
		}

		// NOTE: a minor problem: When there is no internet connection, this code may block
		// forever. Is there a way to solve this?
		// TODO: Yes, timeout
		ip, err := localip.LocalIP()
		if err != nil {
			fmt.Println("Could not get local IP address. Error:", err)
			fmt.Println("Retrying...")
			time.Sleep(time.Second)
			continue
		}

		IP := net.ParseIP(ip)

		ThisNode = newElevator(id, IP, state, localRequestChan)

		break
	}

	go ThisNode.requestListener.Listen()
	<-ThisNode.requestListener.ReadyChan

	fmt.Println("Successfully created new network node: ")
	fmt.Println(ThisNode)

	go transfer.BroadcastSender(stateBroadcastPort, LifeSignalChan)
	go transfer.BroadcastReceiver(stateBroadcastPort, LifeSignalChan)

	go ThisNode.timeout()
	go ThisNode.sendLifeSignal(LifeSignalChan)
	go ThisNode.readLifeSignals(LifeSignalChan)
}

func newElevator(id string, ip net.IP, state *elevalgo.Elevator, localRequestChan chan ElevatorRequest) node {
	return node{
		id:    id,
		state: state,
		ip:    ip,
		requestListener: transfer.NewListener(net.UDPAddr{
			IP:   ip,
			Port: transfer.GetAvailablePort(),
		}),
		localRequestChan: localRequestChan,
		peers:            make([]*peer, 0),
		peersLock:        &sync.Mutex{},
	}
}

func newPeer(requestSender transfer.Sender, state elevalgo.Elevator, id string) *peer {
	return &peer{
		sender: requestSender,
		info:   newPeerInfo(state, id),
	}
}

func newPeerInfo(state elevalgo.Elevator, id string) PeerInfo {
	return PeerInfo{
		State:     state,
		Id:        id,
		LastSeen:  time.Now(),
		Connected: false,
	}
}

func (n node) String() string {
	return fmt.Sprintf("Elevator %s, listening on: %s\n", n.id, &n.requestListener.Addr)
}

func (p peer) String() string {
	return fmt.Sprintf("Peer %s, Sender object:\n %s\n", p.info.Id, p.sender)
}
