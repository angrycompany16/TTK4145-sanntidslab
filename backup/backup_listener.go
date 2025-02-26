package backup

import (
	"fmt"
	elevalgo "sanntidslab/elev_al_go"
	networking "sanntidslab/network"
	"sync"
	"time"

	"github.com/angrycompany16/Network-go/network/localip"
	//elevalgo "sanntidslab/elev_al_go"
	//timer "sanntidslab/elev_al_go/timer"
)

var (
	thisBackup = Backup{
		password: Password,
	}
)

type Backup struct {
	primaryIP string
	password  string
	lastSeen  time.Time
	aliveLock *sync.Mutex
	state     elevalgo.Elevator
}

func HandleLifeSignal(lifesignal networking.LifeSignal) {
	// GOAL OF FUNCTION: if the listener detects that main is ded, revive. if alive update backupView

	// while timer not expired {
	// 1. read lifesignal of targetIP

	// 2.1 if lifesignal not recieved within timeout, revive TargetIP

	// 2.2 if lifesignal recieved, update backupView
	//}

	localIP, err := localip.LocalIP()
	if err != nil {
		fmt.Print("no")
	}

	if lifesignal.ListenerAddr.IP.String() != localIP {
		fmt.Println("Backed up", lifesignal.ListenerAddr.IP.String())
		thisBackup.primaryIP = lifesignal.ListenerAddr.IP.String()
		thisBackup.lastSeen = time.Now()
		thisBackup.state = lifesignal.State
	}
}

func ReviveTimeout() {
	timeout := time.Second * 6

	for {
		if thisBackup.lastSeen.Add(timeout).Before(time.Now()) {
			Revive(thisBackup.primaryIP, thisBackup.password)
		}
	}
}
