package server

import (
	"container/list"
	"github.com/sblundy/inmemorytftp/server/packets"
	"strings"
	"testing"
)

func TestHandleWriteRequest_EmptyFile(t *testing.T) {
	dummyConn := NewDummyPacketConn("TestHandleWriteRequest_EmptyFile", packets.NewData(1, []byte{}))

	HandleWriteRequest(&dummyConn, "test.txt")

	assertNumSent(t, dummyConn.packetWritten, 2)
	assertAckPacket(t, dummyConn.packetWritten.Front(), 0)
	assertAckPacket(t, dummyConn.packetWritten.Back(), 1)
}

func TestHandleWriteRequest_FileOfPacketLength(t *testing.T) {
	dummyConn := NewDummyPacketConn("TestHandleWriteRequest_EmptyFile",
		packets.NewData(1, []byte(strings.Repeat("12345678", 64))),
		packets.NewData(2, []byte{}))

	HandleWriteRequest(&dummyConn, "test.txt")

	assertNumSent(t, dummyConn.packetWritten, 3)
	assertAckPacket(t, dummyConn.packetWritten.Front(), 0)
	assertAckPacket(t, dummyConn.packetWritten.Back(), 2)
}

func assertAckPacket(t *testing.T, actual *list.Element, expectedBlock uint16) {
	switch actual.Value.(type) {
	default:
		t.Error("Incorrect type packet sent", actual.Value)
	case packets.AckPacket:
		dataPacket := actual.Value.(packets.AckPacket)
		if expectedBlock != dataPacket.Block {
			t.Error("Block incorrect in packet", dataPacket.Block)
		}
	}
}
