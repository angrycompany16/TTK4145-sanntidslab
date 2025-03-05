package backup

import (
	"fmt"
	"sync"
	"time"

	networking "sanntidslab/network"

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

// TODO: Connect backup to life signals from network.go

func HandleLifeSignal(lifeSignal networking.LifeSignal) {
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

// funksjon som etterspør backup på ordre
func requestBackup(peerList []networking.Peer) (success bool) {
	// ask a random peer to backup a new order
	// for _, peer := range peerList {
	// Choose first free peer
	// Blocking timeout when we wait for answer
	// if backup received, return true

	// return false
	//}
	return false
}

// funksjon som svarer på backup av ordre
func acceptingBackup(request networking.ElevatorRequest) {
	thisBackup.backupOrders = append(thisBackup.backupOrders, backupOrder{
		id:         request.SenderId,
		lastSeen:   time.Now(),
		buttonType: request.ButtonType,
		floor:      request.Floor,
	})
}

func sendBackup() {
	// send cab orders back to revived elevators
}
