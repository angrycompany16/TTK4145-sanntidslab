package distribute

import (
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"net"
	"reflect"
	elevalgo "sanntidslab/elev_al_go"
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
	ListenerAddr net.UDPAddr
	SenderId     string
	State        elevalgo.Elevator
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
type node struct {
	id               string
	state            *elevalgo.Elevator
	ip               net.IP
	requestListener  transfer.Listener
	peers            []*peer
	peersLock        *sync.Mutex
	BackupAckChan    chan Ack
	localRequestChan chan ElevatorRequest
	// OutwardAckChan   chan Ack
}

// Problem: I want the request sender in the peer struct, but it would be nice to instantiate
// a backup sender / listener at the same time
// Solution: Maybe separate the connection parts from each other so that we can ...
// That's going to be difficult, because the sending/connecting is tightly coupled
// with the peer struct
// Perhaps we can move the peer somewhere else and use only the node structs?
// Not sure, it is quite a strange conundrum
// For now, let's keep them in the same file
type peer struct {
	sender transfer.Sender
	// ackSender    transfer.Sender // Sender through which we send records to be backed up
	// recordSender transfer.Sender // Sender through which we send records to be backed up
	state    elevalgo.Elevator
	id       string
	lastSeen time.Time
}

func (n *node) timeout() {
	for {
		n.peersLock.Lock()
		for i, peer := range n.peers {
			if peer.lastSeen.Add(timeout).Before(time.Now()) {
				fmt.Println("Removing peer:", peer)
				peer.sender.QuitChan <- 1
				n.requestListener.QuitChan <- peer.id
				n.peers[i] = n.peers[len(n.peers)-1]
				n.peers = n.peers[:len(n.peers)-1]
			}
		}
		n.peersLock.Unlock()
	}
}

func (n *node) sendLifeSignal(signalChan chan (LifeSignal)) {
	for {
		signal := LifeSignal{
			ListenerAddr: n.requestListener.Addr,
			SenderId:     n.id,
			State:        *n.state,
		}

		signalChan <- signal
		time.Sleep(time.Millisecond * 10)
	}
}

func (n *node) readLifeSignals(signalChan chan (LifeSignal)) {
LifeSignals:
	for lifeSignal := range signalChan {
		if n.id == lifeSignal.SenderId {
			continue
		}

		n.peersLock.Lock()
		for _, _peer := range n.peers {
			if _peer.id == lifeSignal.SenderId {
				_peer.lastSeen = time.Now()
				_peer.state = lifeSignal.State

				// TODO: Can be made better
				// We want to connect that boy
				if !_peer.sender.Connected {
					go _peer.sender.Send()
					<-_peer.sender.ReadyChan
					_peer.sender.Connected = true
				}

				n.peersLock.Unlock()

				continue LifeSignals
			}
		}

		requestSender := transfer.NewSender(lifeSignal.ListenerAddr, n.id)

		newPeer := newPeer(requestSender, lifeSignal.State, lifeSignal.SenderId)

		// Send backed up requests back to the new elevator

		n.peers = append(n.peers, newPeer)
		fmt.Println("New peer added: ")
		fmt.Println(newPeer)

		n.peersLock.Unlock()
	}
}

// Sends a request given button type and floor to the first free node
// Returns false if the message was sent away, true if it should be handled by this elevator
// TODO: Return value may not be needed here
func (n *node) SendRequest(button elevio.ButtonEvent) bool {
	if button.Button == elevio.BT_Cab {
		return true
	}

	req := ElevatorRequest{
		SenderId:   n.id,
		ButtonType: button.Button,
		Floor:      button.Floor,
	}

	assigneeID := n.getBestElevator()

	// Instead of waiting for some peer to become available, assign to ourselves if there
	// is no peer that responds.
	// NOTE: This may create deadlocking problems
	if assigneeID == n.id {
		// Request a node to backup the request
		// Enter a for loop:
		// Create a record
		// Send the record to the peer
		// Wait for ack
		// Return if timeout
		record := Record{
			Request: req,
			Id:      n.id,
		}
		// NOTE: Right now this simply assumes that the first node will be free. Obviously
		// this is not so great
		for _, peer := range n.peers {
			// Timeout
			peer.sender.DataChan <- newGeneralMsg(record) // Send the backup request
			// If timeout, continue

			// Listen for ack
			// If ack times out, also continue
			acknowledge := <-n.BackupAckChan
			fmt.Println("acknowledge received", acknowledge)
			n.localRequestChan <- req
			return true
			// Send to active requests
			// If timeout, continue
			// else we are backed up, move on out of the loop
		}
		// TODO: make this system
	}

	n.peersLock.Lock()
	for _, peer := range n.peers {
		if peer.id == assigneeID {
			// Backup the request
			ThisBackup.AddRecord(req, peer.id)
			// Send to the peer
			peer.sender.DataChan <- newGeneralMsg(req) // Send the elevator request
			n.peersLock.Unlock()
			return false
		}
	}
	n.peersLock.Unlock()
	return true
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
	elevatorList := []elevalgo.Elevator{*n.state}
	for _, peer := range n.peers {
		elevatorList = append(elevatorList, peer.state)
	}

	var winnerID string
	n.peersLock.Lock() // Changing the peer list during this operation is probably not a good
	orderList := elevalgo.GetBestOrder(elevatorList)
	if orderList[0] == 0 {
		winnerID = n.id
	} else {
		winnerID = n.peers[orderList[0]-1].id
	}
	n.peersLock.Unlock()

	return winnerID
}

func (n *node) SendAck(id string) {
	for _, peer := range n.peers {
		if peer.id == id {
			peer.sender.DataChan <- Ack{}
		}
	}
}

// TODO: Need a way to check which struct was actually received
func (n *node) PipeListener(requestRx chan ElevatorRequest, ackRx chan Ack, recordRx chan Record) {
	for msg := range n.requestListener.DataChan {
		var message GeneralMsg
		// The core of the problem: GeneralMsg has a type interface{}, so it won't
		// be able to encode nested structs as they will just be stored as maps
		// fmt.Println("Message:", msg)
		mapstructure.Decode(msg, &message)
		// Structify(&msg, &message)
		// fmt.Println(message)
		switch message.TypeName {
		case reflect.TypeOf(ElevatorRequest{}).Name():
			var request ElevatorRequest
			err := mapstructure.Decode(message.Data, &request)
			if err != nil {
				log.Fatal("Died")
			}
			requestRx <- request
		case reflect.TypeOf(Ack{}).Name():
			var ack Ack
			err := mapstructure.Decode(message.Data, &ack)
			if err != nil {
				log.Fatal("Died ack og ve")
			}
			ackRx <- ack
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

func (n *node) LocalRequests(requestChan chan ElevatorRequest) {

}

func InitNode(state *elevalgo.Elevator) {
	for {
		var id string
		flag.StringVar(&id, "id", "", "id of this peer")

		flag.Parse()

		if id == "" {
			r := rand.Int()
			fmt.Println("No id was given. Using randomly generated number", r)
			id = strconv.Itoa(r)
		}

		ip, err := localip.LocalIP()
		if err != nil {
			fmt.Println("Could not get local IP address. Error:", err)
			fmt.Println("Retrying...")
			time.Sleep(time.Second)
			continue
		}

		IP := net.ParseIP(ip)

		ThisNode = newElevator(id, IP, state)

		break
	}

	go ThisNode.requestListener.Listen()
	<-ThisNode.requestListener.ReadyChan

	// go ThisNode.ackListener.Listen()
	// <-ThisNode.requestListener.ReadyChan

	// go ThisNode.recordListener.Listen()
	// <-ThisNode.requestListener.ReadyChan

	fmt.Println("Successfully created new network node: ")
	fmt.Println(ThisNode)

	go transfer.BroadcastSender(stateBroadcastPort, LifeSignalChan)
	go transfer.BroadcastReceiver(stateBroadcastPort, LifeSignalChan)

	go ThisNode.timeout()
	go ThisNode.sendLifeSignal(LifeSignalChan)
	go ThisNode.readLifeSignals(LifeSignalChan)
}

func newElevator(id string, ip net.IP, state *elevalgo.Elevator) node {
	return node{
		id:            id,
		state:         state,
		ip:            ip,
		BackupAckChan: make(chan Ack),
		requestListener: transfer.NewListener(net.UDPAddr{
			IP:   ip,
			Port: transfer.GetAvailablePort(),
		}),
		peers:     make([]*peer, 0),
		peersLock: &sync.Mutex{},
	}
}

func newPeer(requestSender transfer.Sender, state elevalgo.Elevator, id string) *peer {
	return &peer{
		sender:   requestSender,
		state:    state,
		id:       id,
		lastSeen: time.Now(),
	}
}

// func NewAck(id string) Ack {
// 	return Ack{id: id}
// }

func (n node) String() string {
	return fmt.Sprintf("Elevator %s, listening on: %s\n", n.id, &n.requestListener.Addr)
}

func (p peer) String() string {
	return fmt.Sprintf("Peer %s, Sender object:\n %s\n", p.id, p.sender)
}
