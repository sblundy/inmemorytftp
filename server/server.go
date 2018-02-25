package server

import (
	"fmt"
	"github.com/sblundy/inmemorytftp/server/connection"
	"github.com/sblundy/inmemorytftp/server/packets"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

type TftpServer struct {
	logger       *log.Logger
	port         uint
	run          bool
	runCheckFreq time.Duration
}

func New(port uint, runCheckFreq time.Duration) TftpServer {
	return TftpServer{
		logger:       log.New(os.Stderr, "TftpServer ", log.LstdFlags),
		port:         port,
		run:          true,
		runCheckFreq: runCheckFreq,
	}
}

func (server *TftpServer) Listen() {
	listenAddr := fmt.Sprintf(":%d", server.port)
	conn, err := net.ListenPacket("udp", listenAddr)
	if err != nil {
		server.logger.Fatalln("Unable to open port", server.port, err)
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
				server.logger.Println(err.Error())
			case *net.OpError:
				opErr := err.(*net.OpError)
				if opErr.Timeout() {
					continue
				} else {
					server.logger.Println(err.Error())
				}
			}
		} else if n == 0 {
			server.logger.Println("Packet is empty", addr)
		} else if n < 2 {
			server.logger.Println("Packet too short", addr)
		} else {
			go server.handlePacket(conn, buff[:n], addr)
		}
	}
}

func (server *TftpServer) Stop() {
	server.run = false
}

func (server *TftpServer) handlePacket(conn net.PacketConn, buff []byte, addr net.Addr) {
	server.logger.Println("Packet received", addr, buff)
	packet, ok := packets.Read(buff)
	initialConnection := connection.WrapExisting(conn, addr)
	if ok {
		switch packet.(type) {
		default:
			server.handleDefault(initialConnection, packet)
		case packets.ReadPacket:
			server.onReadRequest(initialConnection, packet.(packets.ReadPacket), addr)
		case packets.WritePacket:
			server.onWriteRequest(initialConnection, packet.(packets.WritePacket), addr)
		case packets.DataPacket:
			server.onData(initialConnection, packet.(packets.DataPacket))
		case packets.AckPacket:
			server.onAck(initialConnection, packet.(packets.AckPacket))
		case packets.ErrorPacket:
			server.onError(initialConnection, packet.(packets.ErrorPacket))
		}
	}
}

func (server *TftpServer) handleDefault(replyChannel connection.TftpReplyChannel, packet packets.Packet) {
	server.logger.Println("Unexpected packet received", packet)
	replyChannel.Write(packets.NewError(4, "Not understood"))
}

func (server *TftpServer) onReadRequest(replyChannel connection.TftpReplyChannel, packet packets.ReadPacket, target net.Addr) {
	var fileBytes []byte = nil
	//TODO impl
	if packet.Filename == "dummy.txt" {
		fileBytes = []byte(strings.Repeat("1234567890", 100))
	}
	conn, err := connection.New(target)
	if err != nil {
		log.Println("Unable to open a local port!", err)
		replyChannel.Write(packets.NewError(0, "Unable to open local port"))
		return
	}
	defer conn.Close()
	if fileBytes == nil {
		replyChannel.Write(packets.NewError(1, "File not found"))
	} else {
		HandleReadRequest(conn, fileBytes)
	}
}

func (server *TftpServer) onWriteRequest(replyChannel connection.TftpReplyChannel, packet packets.WritePacket, sender net.Addr) {
	server.logger.Println("in onData")
	if len(packet.Filename) == 0 {
		replyChannel.Write(packets.NewError(4, "Zero length file name not allowed"))
		return
	}

	conn, err := connection.New(sender)
	if err != nil {
		log.Println("Unable to open a local port!", err)
		replyChannel.Write(packets.NewError(0, "Unable to open local port"))
		return
	}
	defer conn.Close()

	HandleWriteRequest(conn, packet.Filename)
}

func (server *TftpServer) onData(replyChannel connection.TftpReplyChannel, packet packets.DataPacket) {
	server.logger.Println("in onData")
	replyChannel.Write(packets.NewError(5, "Data not expected"))
}

func (server *TftpServer) onAck(replyChannel connection.TftpReplyChannel, packet packets.AckPacket) {
	server.logger.Println("in onAck")
	//ignoring
}

func (server *TftpServer) onError(replyChannel connection.TftpReplyChannel, packet packets.ErrorPacket) {
	server.logger.Println("in onError")
	//ignoring
}
