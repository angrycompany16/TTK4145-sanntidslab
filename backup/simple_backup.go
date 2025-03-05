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

type Acknowledge struct {
	Ack bool
}

// TODO: Connect backup to life signals from network.go

func HandleLifeSignal() {
	lifeSignal := <-LifeSignalChan

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

	answerChan := make(chan Acknowledge)
	quitChan := make(chan bool)
	defer close(answerChan)
	defer close(quitChan)

	go ThisNode.pipeAcknowledge(answerChan, quitChan)

	success := false
	for _, peer := range ThisNode.peers {
		peer.sender.DataChan <- order // Ask peer for backup

		select { // Wait for acknowledge, if timeout continue and ask next peer
		case <-answerChan:
			quitChan <- true
			return true

		case <-time.After(1 * time.Second):
			continue
		}
	}

	quitChan <- true
	return success
}

func (n *node) pipeAcknowledge(answerChan chan Acknowledge, quitChan chan bool) {
	for {
		select {
		case <-quitChan:
			return

		default:
			for msg := range n.listener.DataChan {
				var a Acknowledge
				n.listener.DecodeMsg(&msg, &a)
			}

			answerChan <- a
		}
	}
}

func (n *node) pipeOrderListener() {
	for msg := range n.listener.DataChan {
		var order backupOrder
		n.listener.DecodeMsg(&msg, &order)
		acceptingBackup(order)

		for _, peer := range n.peers {
			if peer.id == order.id {
				peer.sender.DataChan <- Acknowledge{Ack: true}
			}
		}

	}
}

// funksjon som svarer på backup av ordre
func acceptingBackup(order backupOrder) (success bool) {
	thisBackup.backupOrders = append(thisBackup.backupOrders, order)
	return true

}

// TODO: log all calls done when DC

func sendBackup() {
	// send cab orders back to revived elevators
}

// TODO: take over all hall calls