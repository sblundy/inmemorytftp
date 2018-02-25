package server

import (
	"bytes"
	"container/list"
	"github.com/sblundy/inmemorytftp/server/packets"
	"strings"
	"testing"
)

func TestHandleWriteRequest_EmptyFile(t *testing.T) {
	dummyConn := NewDummyPacketConn("TestHandleWriteRequest_EmptyFile", packets.NewData(1, []byte{}))

	output, ok := HandleWriteRequest(&dummyConn, "test.txt")

	assertSuccess(t, ok, output, []byte{})
	assertNumSent(t, dummyConn.packetWritten, 2)
	assertAckPacket(t, dummyConn.packetWritten.Front(), 0)
	assertAckPacket(t, dummyConn.packetWritten.Back(), 1)
}

func TestHandleWriteRequest_FileOfPacketLength(t *testing.T) {
	fileContents := []byte(strings.Repeat("12345678", 64))
	dummyConn := NewDummyPacketConn("TestHandleWriteRequest_FileOfPacketLength",
		packets.NewData(1, fileContents),
		packets.NewData(2, []byte{}))

	output, ok := HandleWriteRequest(&dummyConn, "test.txt")

	assertSuccess(t, ok, output, fileContents)
	assertNumSent(t, dummyConn.packetWritten, 3)
	assertAckPacket(t, dummyConn.packetWritten.Front(), 0)
	assertAckPacket(t, dummyConn.packetWritten.Back(), 2)
}

func TestHandleWriteRequest_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	fileContents := []byte(strings.Repeat("12345678", 64))
	dummyConn := NewDummyPacketConn("TestHandleWriteRequest_Timeout",
		packets.NewData(1, fileContents))

	_, ok := HandleWriteRequest(&dummyConn, "test.txt")

	if ok {
		t.Error("Expected to fail")
	}

	assertAckPacket(t, dummyConn.packetWritten.Front(), 0)
	assertAckPacket(t, dummyConn.packetWritten.Back(), 1)
}
func TestHandleWriteRequest_ResendsAckOnDuplicateData(t *testing.T) {
	fileContents := []byte(strings.Repeat("12345678", 64))
	dummyConn := NewDummyPacketConn("TestHandleWriteRequest_EmptyFile", packets.NewData(1, fileContents),
		packets.NewData(1, fileContents),
		packets.NewData(2, []byte{}))

	output, ok := HandleWriteRequest(&dummyConn, "test.txt")

	assertSuccess(t, ok, output, fileContents)
	assertNumSent(t, dummyConn.packetWritten, 4)
	assertAckPacket(t, dummyConn.packetWritten.Front(), 0)
	assertAckPacket(t, dummyConn.packetWritten.Back().Prev().Prev(), 1)
	assertAckPacket(t, dummyConn.packetWritten.Back().Prev(), 1)
	assertAckPacket(t, dummyConn.packetWritten.Back(), 2)
}

func assertSuccess(t *testing.T, ok bool, contents []byte, expectedContents []byte) {
	t.Helper()
	if !ok {
		t.Error("Write failed when expected to succeed")
	} else if !bytes.Equal(contents, expectedContents) {
		t.Error("Contents not correct", contents)
	}
}

func assertAckPacket(t *testing.T, actual *list.Element, expectedBlock uint16) {
	t.Helper()
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
