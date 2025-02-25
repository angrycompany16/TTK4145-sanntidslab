package backup

import (
	"flag"
	"sync"
	//elevalgo "sanntidslab/elev_al_go"
	//timer "sanntidslab/elev_al_go/timer"
	//networking "sanntidslab/network"
)

type Backup struct {
	primaryIP string
	password  string
	AliveLock *sync.Mutex
	//BackupView []elevalgo.Elevator                               ----spør david hva faen dette er
}

func backupFunctionality(backup *Backup) {
	// GOAL OF FUNCTION: if the listener detects that main is ded, revive. if alive update backupView

	// while timer not expired {
	// 1. read lifesignal of targetIP

	// 2.1 if lifesignal not recieved within timeout, revive TargetIP

	// 2.2 if lifesignal recieved, update backupView
	//}

}

func main() {
	var node string
	flag.StringVar(&node, "node", "", "flag to be able to tell if program has backup running")

	var hostIP string
	flag.StringVar(&hostIP, "hostIP", hostIP, "Get hostIP from flag")

	flag.Parse()

	backup := Backup{
		primaryIP: hostIP,
		password:  "password",
		AliveLock: &sync.Mutex{},
		//BackupView: []elevalgo.Elevator{},                               ----spør david hva faen dette er
	}

	for {
		backupFunctionality(&backup)
	}
}
