package main

import (
	"flag"
	"fmt"
	"github.com/sblundy/inmemorytftp/server"
	"os"
	"time"
)

func main() {
	opts := flag.NewFlagSet("inmemorytftp", flag.ContinueOnError)
	port := opts.Uint("port", 69, "Port to listen for connections")
	err := opts.Parse(os.Args[1:])
	if err != nil {
		switch err {
		default:
			os.Exit(1)
		case flag.ErrHelp:
			return
		}
	}
	fmt.Printf("Listening on %d\n", *port)
	service := server.New(*port, 10*time.Second)
	service.Listen()
}
