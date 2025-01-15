package main

// server IP address:
// 10.100.23.204:34933

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	// go fixedSizeTCP()
	go nullTerminatedTCP()

	for {

	}
}

func nullTerminatedTCP() {
	server, err := net.ResolveTCPAddr("tcp", "10.100.23.204:33546")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialTCP("tcp", nil, server)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	for {
		data, err := bufio.NewReader(conn).ReadString('\000')
		if err != nil {
			log.Print(err)
			continue
		}

		fmt.Println(data)

		conn.Write([]byte("Hello from group 56 (Canadaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa Canadaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa Canadaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa Canadaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa Canadaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa Canadaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa hello)\000"))

		// conn.Write([]byte("Hello from group 56 (Canada hello)\000"))

		time.Sleep(100 * time.Millisecond)
	}

	// buffer := make([]byte, 1024)

	// for {
	// 	msg := ""

	// readLoop:
	// 	for {
	// 		_, err = conn.Read(buffer)

	// 		if err != nil {
	// 			log.Print(err)
	// 			continue
	// 		}

	// 		fmt.Println("Hello")

	// 		substr := string(buffer)

	// 		for _, char := range substr {
	// 			msg += string(char)
	// 			if char == 0 {
	// 				fmt.Print(msg)
	// 				msg = ""
	// 				break readLoop
	// 			}
	// 		}
	// 	}

	//

	// 	// if err != nil {
	// 	// 	log.Print(err)
	// 	// 	continue
	// 	// }

	// 	time.Sleep(100 * time.Millisecond)
	// }
}

func fixedSizeTCP() {
	server, err := net.ResolveTCPAddr("tcp", "10.100.23.204:34933")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialTCP("tcp", nil, server)
	if err != nil {
		log.Fatal(err)
	}

	defer conn.Close()

	buffer := make([]byte, 1024)

	for {
		_, err = conn.Read(buffer)

		if err != nil {
			log.Print(err)
			continue
		}

		fmt.Println(string(buffer))

		_, err := conn.Write([]byte("Hello from group 56 (Canada hello)\000"))

		if err != nil {
			log.Print(err)
			continue
		}

		time.Sleep(time.Millisecond * 100)
	}
}
