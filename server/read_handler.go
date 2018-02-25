package server

import (
	"fmt"
	"github.com/sblundy/inmemorytftp/server/connection"
	"github.com/sblundy/inmemorytftp/server/packets"
	"log"
	"os"
	"time"
)

const MaxPayloadSize = 512
const maxBlockRetries = 3

func HandleReadRequest(conn connection.TftpPacketConn, payload []byte) {
	logger := log.New(os.Stdout, fmt.Sprintf("TftpServer.ReadRequest(%s->%s) ", conn.LocalAddr(), conn.RemoteAddr()), log.LstdFlags)
	logger.Println("Start read")
	var numBlocks = (uint16(len(payload) / MaxPayloadSize)) + 1
	for blockId := uint16(1); blockId < numBlocks; blockId++ {
		startIndex := (blockId - 1) * MaxPayloadSize
		nextStartIndex := blockId * MaxPayloadSize
		block := payload[startIndex:nextStartIndex]
		if !sendBlock(conn, blockId, block, logger) {
			conn.Write(packets.NewError(5, "Send failed"))
			logger.Println("End send:failed")
			return
		}
	}
	lastBlockStartIndex := (numBlocks - 1) * MaxPayloadSize
	if !sendBlock(conn, numBlocks, payload[lastBlockStartIndex:], logger) {
		conn.Write(packets.NewError(5, "Send failed"))
		logger.Println("End send:failed")
	} else {
		logger.Println("End send")
	}
}

func sendBlock(conn connection.TftpPacketConn, blockId uint16, block []byte, logger *log.Logger) bool {
	if trySendBlock(conn, blockId, block, logger) {
		return true
	} else {
		for i := 0; i < maxBlockRetries; i++ {
			logger.Printf("Failed to send data. Retry %d/%d", i+1, maxBlockRetries)
			if trySendBlock(conn, blockId, block, logger) {
				return true
			}
		}
		return false
	}
}

func trySendBlock(conn connection.TftpPacketConn, blockId uint16, block []byte, logger *log.Logger) bool {
	packet := packets.NewData(blockId, block)
	ok := conn.Write(packet)
	if !ok {
		return false
	}

	if !receiveAck(conn, blockId, logger) {
		logger.Println("Ack not received")
		return false
	}
	return true
}

func receiveAck(conn connection.TftpPacketConn, block uint16, logger *log.Logger) bool {
	packet, ok := conn.Read(10 * time.Second)
	if !ok {
		return false
	}

	switch packet.(type) {
	default:
		logger.Println("Unexpected packet received")
	case packets.AckPacket:
		ack := packet.(packets.AckPacket)
		if ack.Block == block {
			return true
		}
	}

	return false
}
