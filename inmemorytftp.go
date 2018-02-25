package main

import (
	"flag"
	"fmt"
	"github.com/sblundy/inmemorytftp/server"
	"time"
)

func main() {
	port := flag.Uint("port", 69, "Port to listen for connections")
	flag.Parse()
	fmt.Printf("Listening on %d\n", *port)
	service := server.New(*port, 10*time.Second)
	service.Listen()
}
