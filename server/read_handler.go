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
const readBlockTimeout = 30 * time.Second

func HandleReadRequest(conn connection.TftpPacketConn, payload []byte) {
	logger := log.New(os.Stdout, fmt.Sprintf("TftpServer.ReadRequest(%s->%s) ", conn.LocalAddr(), conn.RemoteAddr()), log.LstdFlags)
	logger.Println("Start read")
	var numBlocks = (uint16(len(payload) / MaxPayloadSize)) + 1
	for blockId := uint16(1); blockId < numBlocks; blockId++ {
		startIndex := (blockId - 1) * MaxPayloadSize
		nextStartIndex := blockId * MaxPayloadSize
		block := payload[startIndex:nextStartIndex]
		if !sendBlock(conn, blockId, block, logger) {
			logger.Println("ERROR: End send:failed")
			return
		}
	}
	lastBlockStartIndex := (numBlocks - 1) * MaxPayloadSize
	if !sendBlock(conn, numBlocks, payload[lastBlockStartIndex:], logger) {
		logger.Println("ERROR: End send:failed")
	} else {
		logger.Println("End send")
	}
}

func sendBlock(conn connection.TftpPacketConn, blockId uint16, block []byte, logger *log.Logger) bool {
	deadline := time.Now().Add(readBlockTimeout)
	retry := 0
	for time.Now().Before(deadline) {
		result := trySendBlock(conn, blockId, block, logger)
		switch result {
		case AckNotReceived:
			retry++
		case AckReceived:
			return true
		case WriteFailed:
			conn.Write(packets.NewError(5, "Send failed"))
			return false
		case PrematureTermination:
			return false
		}
	}

	conn.Write(packets.NewError(5, "Send failed"))
	return false
}

func trySendBlock(conn connection.TftpPacketConn, blockId uint16, block []byte, logger *log.Logger) responseType {
	packet := packets.NewData(blockId, block)
	ok := conn.Write(packet)
	if !ok {
		return WriteFailed
	}

	return receiveAck(conn, blockId, logger)
}

type responseType int

const (
	_              = iota
	AckNotReceived = iota
	AckReceived
	WriteFailed
	PrematureTermination
)

func receiveAck(conn connection.TftpPacketConn, block uint16, logger *log.Logger) responseType {
	packet, ok := conn.Read(10 * time.Second)
	if !ok {
		return AckNotReceived
	}

	switch packet.(type) {
	default:
		logger.Println("WARN: Unexpected packet received")
	case packets.ErrorPacket:
		return PrematureTermination
	case packets.AckPacket:
		ack := packet.(packets.AckPacket)
		if ack.Block == block {
			return AckReceived
		}
	}

	return AckNotReceived
}
