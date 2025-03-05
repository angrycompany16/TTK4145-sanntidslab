package networking

import (
	"flag"
	"fmt"
	"math/rand/v2"
	"net"
	elevalgo "sanntidslab/elev_al_go"
	"strconv"
	"sync"
	"time"

	"github.com/angrycompany16/Network-go/network/localip"
	"github.com/angrycompany16/Network-go/network/transfer"
	"github.com/angrycompany16/driver-go/elevio"
	"github.com/eiannone/keyboard"
)

// NOTE:
// Read this if the buffer size warning appears
// https://github.com/quic-go/quic-go/wiki/UDP-Buffer-Sizes
// TL;DR
// Run
// sudo sysctl -w net.core.rmem_max=7500000
// and
// sudo sysctl -w net.core.wmem_max=7500000

type NodeState int

const (
	Disconnected NodeState = iota
	Connected
)

const (
	stateBroadcastPort = 36251 // Akkordrekke
)

var (
	timeout        = time.Second * 5
	ThisNode       node
	LifeSignalChan = make(chan LifeSignal)
)

type LifeSignal struct {
	ListenerAddr net.UDPAddr
	SenderId     string
	State        elevalgo.Elevator
	WorldView    []elevalgo.Elevator
}

type ElevatorRequest struct {
	SenderId   string
	ButtonType elevio.ButtonType
	Floor      int
}

type node struct {
	id        string
	state     *elevalgo.Elevator
	ip        net.IP
	listener  transfer.Listener
	peers     []*Peer
	peersLock *sync.Mutex
}

type Peer struct {
	sender   transfer.Sender
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
				n.listener.QuitChan <- peer.id
				n.peers[i] = n.peers[len(n.peers)-1]
				n.peers = n.peers[:len(n.peers)-1]
			}
		}
		n.peersLock.Unlock()
	}
}

func (n *node) GetDebugInput() bool {
	// Debug input
	char, key, err := keyboard.GetSingleKey()
	if err != nil {
		fmt.Println("Error when getting key:", err)
		return false
	}

	n.peersLock.Lock()
	if char == 'C' || char == 'c' {
		if len(n.peers) == 0 {
			fmt.Println("No peers!")
		}

		n.SendMsg(elevio.BT_Cab, 3)
	}
	n.peersLock.Unlock()

	return key == keyboard.KeyCtrlC
}

func (n *node) sendLifeSignal(signalChan chan (LifeSignal)) {
	for {
		signal := LifeSignal{
			ListenerAddr: n.listener.Addr,
			SenderId:     n.id,
			State:        *n.state,
		}

		for _, peer := range n.peers {
			signal.WorldView = append(signal.WorldView, peer.state)
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
				// I think QUIC might be the best thing to have graced the earth with its existence
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

		sender := transfer.NewSender(lifeSignal.ListenerAddr, n.id)

		newpeer := newpeer(sender, lifeSignal.State, lifeSignal.SenderId)

		n.peers = append(n.peers, newpeer)
		fmt.Println("New peer added: ")
		fmt.Println(newpeer)

		n.peersLock.Unlock()
	}
}

// Sends a request given button type and floor to the first free node
// Returns false if the message was sent away, true if it should be handled by this elevator
func (n *node) SendMsg(buttonType elevio.ButtonType, floor int) bool {
	req := ElevatorRequest{
		SenderId:   n.id,
		ButtonType: buttonType,
		Floor:      floor,
	}
	for _, peer := range n.peers {
		if isAvailable(&peer.state, &req) {
			peer.sender.DataChan <- req
			return false
		}
	}
	return true
}

func (n *node) PipeListener(receiver chan ElevatorRequest) {
	for msg := range n.listener.DataChan {
		var request ElevatorRequest
		n.listener.DecodeMsg(&msg, &request)
		fmt.Printf("Received request on floor %d, buttontype %d from elevator %s\n", request.Floor, request.ButtonType, request.SenderId)
	}
}

func isAvailable(elevator *elevalgo.Elevator, request *ElevatorRequest) bool {
	// Assume every peer is available
	return true
}

func InitElevator(state *elevalgo.Elevator) {
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

	go ThisNode.listener.Listen()
	<-ThisNode.listener.ReadyChan

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
		id:    id,
		state: state,
		ip:    ip,
		listener: transfer.NewListener(net.UDPAddr{
			IP:   ip,
			Port: transfer.GetAvailablePort(),
		}),
		peers:     make([]*Peer, 0),
		peersLock: &sync.Mutex{},
	}
}

func newpeer(sender transfer.Sender, state elevalgo.Elevator, id string) *Peer {
	return &Peer{
		sender:   sender,
		state:    state,
		id:       id,
		lastSeen: time.Now(),
	}
}

func (n node) String() string {
	return fmt.Sprintf("Elevator %s, listening on: %s\n", n.id, &n.listener.Addr)
}

func (p Peer) String() string {
	return fmt.Sprintf("peer %s, Sender object:\n %s\n", p.id, p.sender)
}
