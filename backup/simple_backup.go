package backup

import (
	"fmt"
	"sync"
	"time"

	// networking "sanntidslab/network"

	"github.com/angrycompany16/driver-go/elevio"
)

var (
	thisBackup = Backup{}
)

type Backup struct {
	backupOrders []backupOrder
	aliveLock    *sync.Mutex
}

type backupOrder struct {
	id         string
	lastSeen   time.Time
	buttonType elevio.ButtonType
	floor      int
}

type acknowledge bool


func BackupFSM() {
	backupRequestChan := make(chan backupOrder)
	orderChan := make(chan ElevatorRequest)
	
	go ThisNode.pipeOrderListener(backupRequestChan)
	go ThisNode.readLifeSignals(LifeSignalChan)

	for {
		select {
		case backupRequest := <-backupRequestChan:
			acception := acceptingBackup(backupRequest)
			receiverId := backupRequest.id
			ThisNode.sendAcknowledge(acception, receiverId)

		case lifeSignal := <-LifeSignalChan:
			updateBackup(lifeSignal)

		case order := <- orderChan:
			requestBackup(order) //success is returned, should we do something with it?
		
		
		}
	}
}

// TODO: Connect backup to life signals from network.go
func updateBackup(lifeSignal LifeSignal) {

	thisBackup.aliveLock.Lock()
	defer thisBackup.aliveLock.Unlock()

	fmt.Println("Handling life signal")

	for i := len(thisBackup.backupOrders) - 1; i >= 0; i-- { // iterate backwards to avoid skipping index when modifying slice
		order := thisBackup.backupOrders[i]
		if order.id == lifeSignal.SenderId {

			//Remove completed orders
			if !lifeSignal.State.Requests[order.floor][order.buttonType] {
				thisBackup.backupOrders = append(thisBackup.backupOrders[:i], thisBackup.backupOrders[i+1:]...)

				//Update lastSeen
			} else {
				thisBackup.backupOrders[i].lastSeen = time.Now()
			}
		}
	}
}

func requestBackup(request ElevatorRequest) bool {
	order := backupOrder{
		id:         request.SenderId,
		lastSeen:   time.Now(),
		buttonType: request.ButtonType,
		floor:      request.Floor,
	}

	answerChan := make(chan acknowledge)
	doneChan := make(chan bool)
	defer close(answerChan)
	defer close(doneChan)

	go ThisNode.pipeAcknowledge(answerChan, doneChan)

	success := false
	for _, peer := range ThisNode.peers {
		peer.sender.DataChan <- order // Ask peer for backup

		select { // Wait for acknowledge, if timeout continue and ask next peer
		case <-answerChan:
			doneChan <- true
			return true

		case <-time.After(1 * time.Second):
			continue
		}
	}

	doneChan <- true
	return success
}

func (n *node) pipeAcknowledge(answerChan chan acknowledge, doneChan chan bool) {
	for {
		select {
		case <-doneChan:
			return

		default:
			for msg := range n.listener.DataChan {
				var a acknowledge
				n.listener.DecodeMsg(&msg, &a)
				answerChan <- a
			}

			
		}
	}
}

func (n *node) pipeOrderListener(orderChan chan backupOrder) {
	for msg := range n.listener.DataChan {
		var order backupOrder
		n.listener.DecodeMsg(&msg, &order)
		orderChan <- order
	}
}

func (n *node) sendAcknowledge(success acknowledge, id string) {
	for _, peer := range n.peers {
		if peer.id == id {
			peer.sender.DataChan <- success
		}
	}
}

// funksjon som svarer på backup av ordre
func acceptingBackup(order backupOrder) (success acknowledge) {
	thisBackup.backupOrders = append(thisBackup.backupOrders, order)
	return true

}

// TODO: log all calls done when DC

func sendBackup() {
	// send cab orders back to revived elevators
}

// TODO: take over all hall calls