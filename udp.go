package main

// server IP address:
// 10.100.23.204:39374
// 10.100.23.24:33184

import (
	"fmt"
	"log"
	"net"
	"time"
)

func listen() {
	address := net.UDPAddr{
		IP:   nil,
		Port: 20013,
	}

	conn, err := net.ListenUDP("udp", &address)

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	buffer := make([]byte, 1024)

	fmt.Println("Start listening")
	for {
		n, _, err := conn.ReadFromUDP(buffer) // n is length of incoming data
		if err != nil {
			log.Print(err)
			continue
		}

		fmt.Print(string(buffer[:n]))
		time.Sleep(time.Millisecond * 100)
	}
}

func send() {
	address := net.UDPAddr{
		IP:   net.ParseIP("10.100.23.204"),
		Port: 20013,
	}

	conn, err := net.DialUDP("udp", nil, &address)

	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	for {
		conn.Write([]byte("Hello\n"))
		time.Sleep(time.Millisecond * 100)
	}
}
