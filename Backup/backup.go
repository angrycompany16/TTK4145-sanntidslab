package backup

import (
	"fmt"
	"net"
	
)

var (
	ipaddress  string = "10.100.23.24"
	password   string = "Sanntid15"
	backupFlag        = "--node=backup"
)

// Plan :- someone asks node, can you be backup?
// node checks if backup listener already created for that IP,
//  if yes return yeah i am backup
//  if not create the backup THEN send yeah i am backup

// GET LOCAL IP UNTILL BETTER SOLUTION IS PRESENTED
func GetLocalIP() (string, error) {
	// Dial UDP to an external address – no data is sent.
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	// Get the local address of the connection.
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func CallBackup(TargetIP string, password string) (backupCreated bool) {

	// Check if the backup is already running on hostPC, and if not start it up
	exitCode, err := AlreadyRunning(backupFlag, ipaddress, password)
	if err != nil {
		fmt.Printf("Failed to Query %v", err)
		backupCreated = false

	} else if exitCode == 0 {
		fmt.Println("Elevatorserver is already running, do nothing")
		backupCreated = true

	} else {
		localIP, err := GetLocalIP()
		if err != nil {
			fmt.Printf("Failed to get local IP %v", err)
			backupCreated = false

		} else {
			CreateBackupListener(TargetIP, localIP, password)
			backupCreated = true
		}
	}
	return backupCreated
}

// if we want to check existence of backup for something
func checkExistence(BackupIP string, password string) (backupExistence bool) {

	// SSH direkte inn og sjekk selv, eller backup life signal av et eller annet slag?

	// Check if backup exists
	exitCodes, err := AlreadyRunning(backupFlag, BackupIP, password)
	if err != nil {
		fmt.Printf("Failed to query: %v", err) // failed to query is same as cannot be backup / DC'd ?
		backupExistence = false

	} else if exitCodes == 0 {
		fmt.Println("Backup is already running, (do nothin)")
		backupExistence = true

	} else {
		fmt.Println("Backup is not running")
		backupExistence = false
	}

	return backupExistence
}