package server

import (
	"bytes"
	"fmt"
	"github.com/sblundy/inmemorytftp/server/connection"
	"github.com/sblundy/inmemorytftp/server/packets"
	"log"
	"os"
	"time"
)

const writeBlockTimeout = 30 * time.Second

func HandleWriteRequest(conn connection.TftpPacketConn, filename string) ([]byte, bool) {
	logger := log.New(os.Stdout, fmt.Sprintf("TftpServer.WriteRequest(%s->%s) ", conn.RemoteAddr(), conn.LocalAddr()), log.LstdFlags)
	logger.Println("Start write", filename)
	conn.Write(packets.NewAck(0))
	buff := bytes.NewBuffer([]byte{})
	var block uint16 = 1
	nextBlockDeadline := time.Now().Add(writeBlockTimeout)
	for time.Now().Before(nextBlockDeadline) {
		switch readPacket(buff, conn, block) {
		case NormalTermination:
			logger.Println("End write", filename, len(buff.Bytes()))
			return buff.Bytes(), true
		case PrematureTerminate:
			logger.Println("WARN: Write terminated", filename)
			return nil, false
		case BlockReceived:
			block++
			// Each time a block is received, update the timeout
			nextBlockDeadline = time.Now().Add(writeBlockTimeout)
		case BlockBotReceived:

		}
	}

	logger.Println("ERROR: End write:timed out", filename)
	return nil, false
}

type readOutcome int

const (
	_                             = iota
	NormalTermination readOutcome = iota
	PrematureTerminate
	BlockReceived
	BlockBotReceived
)

func readPacket(buff *bytes.Buffer, conn connection.TftpPacketConn, block uint16) readOutcome {
	packet, ok := conn.Read(2 * time.Second)
	if !ok {
		//Re-acknowledging the previous block in case that ACK was lost
		conn.Write(packets.NewAck(block - 1))
		return BlockBotReceived
	}
	switch packet.(type) {
	case packets.ErrorPacket:
		return PrematureTerminate
	case packets.DataPacket:
		data := packet.(packets.DataPacket)
		if block == data.Block {
			buff.Write(data.Data)
			conn.Write(packets.NewAck(block))
			if len(data.Data) < MaxPayloadSize {
				//All done
				return NormalTermination
			}
			return BlockReceived
		} else if data.Block < block {
			//Re-acknowledging the previous block in case that ACK was lost
			conn.Write(packets.NewAck(block - 1))
		}
	}
	return BlockBotReceived
}
