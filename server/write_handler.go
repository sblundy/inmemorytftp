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

func HandleWriteRequest(conn connection.TftpPacketConn, filename string) {
	logger := log.New(os.Stdout, fmt.Sprintf("TftpServer.WriteRequest(%s->%s) ", conn.RemoteAddr(), conn.LocalAddr()), log.LstdFlags)
	logger.Println("Start write", filename)
	conn.Write(packets.NewAck(0))
	buff := bytes.NewBuffer([]byte{})
	var block uint16 = 1
	nextBlockDeadline := time.Now().Add(writeBlockTimeout)
	for time.Now().Before(nextBlockDeadline) {
		nextBlock, done := readPacket(buff, conn, block)
		if done {
			logger.Println("End write", filename, buff.Bytes())
			return
		} else if nextBlock {
			block++
			// Each time a block is received, update the timeout
			nextBlockDeadline = time.Now().Add(writeBlockTimeout)
		}
	}

	logger.Println("End write:timed out", filename)
}

func readPacket(buff *bytes.Buffer, conn connection.TftpPacketConn, block uint16) (blockReceived bool, done bool) {
	packet, ok := conn.Read(2 * time.Second)
	if !ok {
		//Re-acknowledging the previous block in case that ACK was lost
		conn.Write(packets.NewAck(block - 1))
		return false, false
	}
	switch packet.(type) {
	case packets.DataPacket:
		data := packet.(packets.DataPacket)
		if block == data.Block {
			buff.Write(data.Data)
			conn.Write(packets.NewAck(block))
			if len(data.Data) < MaxPayloadSize {
				//All done
				return true, true
			}
			return true, false
		} else {
			//Re-acknowledging the previous block in case that ACK was lost
			conn.Write(packets.NewAck(block - 1))
		}
	}
	return false, false
}
