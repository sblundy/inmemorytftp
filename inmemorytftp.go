package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

func main() {
	port := flag.Int("port", 69, "Port to listen for connections")
	flag.Parse()
	fmt.Printf("Listening on %d\n", *port)
	listen(*port)
}

func listen(port int) {
	listenAddr := fmt.Sprintf(":%d", port)
	conn, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		log.Fatalf(err.Error())
		return
	}
	defer conn.Close()

	for {
		buff := make([]byte, 1024)

		n, addr, err := conn.ReadFrom(buff)
		if err != nil {
			log.Fatalf(err.Error())
			return
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
