package main

// Make system detect if an already existing program exists then terminate itself

// Make error handling as uniform as possible, a default function for error handling should take care of MANY errors at once

// Ring structure, everybody needs backup ASAP. not everybodt needs to be alive ASAP (reviving can wait a little while if needed)

// DC and reconnect cannot be a failure, and one elevator cannot fail more than once.

// If a node fails, the backup should be able to revive and SEND BACK the nodes work.

import (
	"bufio"
	"log"
	"net"
)

var (
	backupFlag  string       = "--backup"
	addressPort string       = "localhost:42069" // Store all adresses in advance?
	localAddres *net.TCPAddr = nil
)

func ListenPeer() {
	// listen to peer should be taken from the network module
	return
}

func handleBackupRequest(conn net.Conn) {
	// todo setup the backup procedure
	return
}

func requestBackup(remoteIP string) (bool, error) {

	// establish connecton

	remoteAddr, err := net.ResolveTCPAddr("tcp", remoteIP)
	if err != nil {
		log.Fatal("Error resolving remote address:", err)
	}

	conn, err := net.DialTCP("tcp", localAddres, remoteAddr)
	if err != nil {
		return false, err
	}

	defer conn.Close()

	// Send request for backup, dont know how yet

	backupMessage := []byte("backup_request\n")

	_, err = conn.Write(backupMessage)
	if err != nil {
		return false, err
	}

	response, err := bufio.NewReader(conn).ReadString('\n') // A Node should be able to respond with unavalible?
	if err != nil {
		return false, err
	}

	if response == "backup_accepted\n" {
		// do backup shit
	}

	if response == "backup_rejected\n" {
		// oh fuck oh shit i need backup
	}

	return false, nil
}

func backupData(State string) {
	// do sending of data, dont turn on light before confirmed backup
}
