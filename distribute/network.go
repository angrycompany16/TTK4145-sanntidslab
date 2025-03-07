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

	"github.com/angrycompany16/Network-go/network/broadcast"
	"github.com/angrycompany16/Network-go/network/connection"
	"github.com/angrycompany16/Network-go/network/localip"
	"github.com/angrycompany16/driver-go/elevio"
	"github.com/mitchellh/mapstructure"
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
// - Disconnect check
// - Activity logging

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

// TODO: Write getters

// TODO: Make debugging listeners / Network-go easier
// Should be possible to customize a little
// Add requests as its own struct

type node struct {
	Id                 string
	state              *elevalgo.Elevator
	ip                 net.IP
	requestListener    *connection.Listener
	peers              []*peer
	peersLock          *sync.Mutex
	localRequestChan   chan ElevatorRequest
	unservicedRequests []ElevatorRequest
}

type PeerInfo struct {
	State     elevalgo.Elevator
	Id        string
	LastSeen  time.Time
	Connected bool
}

type peer struct {
	sender connection.Sender
	info   PeerInfo
}

func (n *node) timeout() {
	for {
		n.peersLock.Lock()
		for _, peer := range n.peers {
			if peer.info.LastSeen.Add(timeout).Before(time.Now()) && peer.info.Connected {
				fmt.Println("Lost peer:", peer)
				peer.info.Connected = false
				peer.sender.QuitChan <- 1
				n.requestListener.LostPeers[peer.info.Id] = true
			}
		}
		n.peersLock.Unlock()
	}
}

func (n *node) sendLifeSignal(signalChan chan (LifeSignal)) {
	for {
		// TODO: move this out into its own function
		peerInfoList := make([]PeerInfo, 0)
		for _, peer := range n.peers {
			peerInfoList = append(peerInfoList, peer.info)
		}
		signal := LifeSignal{
			ListenerAddr:       n.requestListener.Addr,
			SenderId:           n.Id,
			State:              *n.state,
			UnservicedRequests: n.unservicedRequests,
			WorldView:          peerInfoList,
		}

		signalChan <- signal
		time.Sleep(time.Millisecond * 10)
	}
}

func (n *node) readLifeSignals(signalChan chan (LifeSignal)) {
LifeSignals:
	for lifeSignal := range signalChan {
		if n.Id == lifeSignal.SenderId {
			continue
		}

		n.peersLock.Lock()
		for _, _peer := range n.peers {
			if _peer.info.Id == lifeSignal.SenderId {
				_peer.info.LastSeen = time.Now()
				_peer.info.State = lifeSignal.State

				// IDEA: Somehow the problem I had solved itself
				if !_peer.info.Connected {
					n.ConnectPeer(_peer, lifeSignal)
					n.peersLock.Unlock()
					continue LifeSignals
				}

				if _peer.sender.Addr.Port != lifeSignal.ListenerAddr.Port {
					fmt.Printf("Sending to port %d, but peer is listening on port %d, making new sender...\n", _peer.sender.Addr.Port, lifeSignal.ListenerAddr.Port)
					_peer.sender.QuitChan <- 1
					n.ConnectPeer(_peer, lifeSignal)
				}

				for _, req := range lifeSignal.UnservicedRequests {
					_peer.info.State.Requests[req.Floor][req.ButtonType] = true
				}

				// Check if our unserviced requests have been backed up
				n.CheckUnservicedRequests(lifeSignal)

				n.peersLock.Unlock()

				continue LifeSignals
			}
		}

		requestSender := connection.NewSender(lifeSignal.ListenerAddr, n.Id)

		newPeer := newPeer(requestSender, lifeSignal.State, lifeSignal.SenderId)

		n.peers = append(n.peers, newPeer)
		fmt.Println("New peer added: ")
		fmt.Println(newPeer)

		n.GetLostRequests(lifeSignal)

		n.peersLock.Unlock()
	}
}

func (n *node) ConnectPeer(_peer *peer, lifeSignal LifeSignal) {
	_peer.sender.Addr = lifeSignal.ListenerAddr
	_peer.sender.Init()
	go _peer.sender.Send()
	_peer.info.Connected = true
}

func (n *node) GetLostRequests(lifeSignal LifeSignal) {
	for _, _peer := range lifeSignal.WorldView {
		if _peer.Id != n.Id {
			continue
		}

		for i := 0; i < elevalgo.NumFloors; i++ {
			for j := 0; j < elevalgo.NumButtons; j++ {
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
		if _peer.Id == n.Id {
			for i, req := range n.unservicedRequests {
				if _peer.State.Requests[req.Floor][req.ButtonType] {
					// TODO: Maybe check if every peer has backed up the request
					n.localRequestChan <- req
					// NOTE: This will reorder the requests
					n.unservicedRequests[i] = n.unservicedRequests[len(n.unservicedRequests)-1]
					n.unservicedRequests = n.unservicedRequests[:len(n.unservicedRequests)-1]
				}
			}
		}
	}
}

// TODO: Learn more about go's concurrency patterns and reconsider the peersLock idea
func (n *node) TakeRequest(request ElevatorRequest, prioritizeSelf bool) {
	if prioritizeSelf {
		n.unservicedRequests = append(n.unservicedRequests, request)
		return
	}

	assigneeID := n.getBestElevator()

	if assigneeID == n.Id {
		if slices.Contains(n.unservicedRequests, request) {
			return
		}
		n.unservicedRequests = append(n.unservicedRequests, request)
	} else {
		// Send the request off to the peer
		for _, _peer := range n.peers {
			if _peer.info.Id == assigneeID {
				_peer.sender.DataChan <- request
			}
		}
	}
}

func (n *node) SelfRequestNode(button elevio.ButtonEvent) ElevatorRequest {
	req := ElevatorRequest{
		SenderId:   n.Id,
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
	if n.peers != nil {
		return n.peers[0].info.Id
	} else {
		return n.Id
	}
}

// TODO: Can simplify a lot
func (n *node) PipeListener(requestRx chan ElevatorRequest) {
	for data := range n.requestListener.DataChan {
		var message connection.Message
		mapstructure.Decode(data, &message)

		switch message.TypeName {
		case reflect.TypeOf(ElevatorRequest{}).Name():
			var request ElevatorRequest
			err := mapstructure.Decode(message.Data, &request)
			if err != nil {
				fmt.Println("Error when decoding elevator request:", err)
				fmt.Println("Request was", data)
				continue
			}
			requestRx <- request
		}
	}
}

func (n *node) GetPeerList() []PeerInfo {
	list := make([]PeerInfo, 0)
	for _, _peer := range n.peers {
		list = append(list, _peer.info)
	}
	return list
}

func InitNode(state *elevalgo.Elevator, localRequestChan chan ElevatorRequest, id string) {
	if id == "" {
		r := rand.Int()
		fmt.Println("No id was given. Using randomly generated number", r)
		id = strconv.Itoa(r)
	}

	ip, err := localip.LocalIP()
	if err != nil {
		log.Fatal("Could not get local IP address. Error:", err)
	}

	IP := net.ParseIP(ip)

	ThisNode = newElevator(id, IP, state, localRequestChan)

	ThisNode.requestListener.Init()
	go ThisNode.requestListener.Listen()

	fmt.Println("Successfully created new network node: ")
	fmt.Println(ThisNode)

	go broadcast.BroadcastSender(stateBroadcastPort, LifeSignalChan)
	go broadcast.BroadcastReceiver(stateBroadcastPort, LifeSignalChan)

	go ThisNode.timeout()
	go ThisNode.sendLifeSignal(LifeSignalChan)
	go ThisNode.readLifeSignals(LifeSignalChan)
}

func newElevator(id string, ip net.IP, state *elevalgo.Elevator, localRequestChan chan ElevatorRequest) node {
	return node{
		Id:    id,
		state: state,
		ip:    ip,
		requestListener: connection.NewListener(net.UDPAddr{
			IP:   ip,
			Port: connection.GetAvailablePort(),
		}),
		localRequestChan: localRequestChan,
		peers:            make([]*peer, 0),
		peersLock:        &sync.Mutex{},
	}
}

func newPeer(requestSender connection.Sender, state elevalgo.Elevator, id string) *peer {
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
	return fmt.Sprintf("Elevator %s, listening on: %s\n", n.Id, &n.requestListener.Addr)
}

func (p peer) String() string {
	return fmt.Sprintf("------- Peer ----\n ~ id: %s\n ~ sends to: %s\n", p.info.Id, &p.sender.Addr)
}
