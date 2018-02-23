package server

import (
	"fmt"
	"log"
	"net"
	"time"
)

type TftpServer struct {
	port         int
	run          bool
	runCheckFreq time.Duration
}

func New(port int, runCheckFreq time.Duration) TftpServer {
	return TftpServer{port: port, run: true, runCheckFreq: runCheckFreq}
}

func (server *TftpServer) Listen() {
	listenAddr := fmt.Sprintf(":%d", server.port)
	conn, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
	defer conn.Close()

	for server.run {
		buff := make([]byte, 1024)

		conn.SetReadDeadline(time.Now().Add(server.runCheckFreq))

		n, addr, err := conn.ReadFrom(buff)
		if err != nil {
			switch err.(type) {
			default:
				log.Fatalf(err.Error())
				return
			case *net.OpError:
				continue
			}
		} else if n == 0 {
			log.Println("Nothing to read")
		} else {
			fmt.Printf("Read successful %d [", n)

			for _, b := range buff[:n] {
				fmt.Printf("%d,", b)
			}
			fmt.Println("]")
		}

		conn.WriteTo([]byte{0, 5, 0, 0, 0}, addr)
	}
}

func (server *TftpServer) Stop() {
	server.run = false
}
