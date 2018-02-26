package server

import (
	"bytes"
	"container/list"
	"github.com/sblundy/inmemorytftp/server/packets"
	"strings"
	"testing"
	"time"
)

func TestHandleReadRequest_EmptyFile(t *testing.T) {
	dummyConn := NewDummyPacketConn("TestHandleReadRequest_EmptyFile", packets.NewAck(1))

	HandleReadRequest(&dummyConn, []byte{})

	assertNumSent(t, dummyConn.packetWritten, 1)
	assertDataPacket(t, dummyConn.packetWritten.Front(), 1, []byte{})
}

func TestHandleReadRequest_OneUnderPacketSize(t *testing.T) {
	dummyConn := NewDummyPacketConn("TestHandleReadRequest_OneUnderPacketSize", packets.NewAck(1))
	file := strings.Repeat("1", MaxPayloadSize-1)

	HandleReadRequest(&dummyConn, []byte(file))

	assertNumSent(t, dummyConn.packetWritten, 1)
	assertDataPacket(t, dummyConn.packetWritten.Front(), 1, []byte(file))
}

func TestHandleReadRequest_PacketSize(t *testing.T) {
	dummyConn := NewDummyPacketConn("TestHandleReadRequest_PacketSize", packets.NewAck(1), packets.NewAck(2))
	file := strings.Repeat("1", MaxPayloadSize)

	HandleReadRequest(&dummyConn, []byte(file))

	assertNumSent(t, dummyConn.packetWritten, 2)
	assertDataPacket(t, dummyConn.packetWritten.Front(), 1, []byte(file))
	assertDataPacket(t, dummyConn.packetWritten.Back(), 2, []byte{})
}

func TestHandleReadRequest_Retry(t *testing.T) {
	dummyConn := NewDummyPacketConn("TestHandleReadRequest_Retry", nil, packets.NewAck(1))

	HandleReadRequest(&dummyConn, []byte{})

	assertNumSent(t, dummyConn.packetWritten, 2)
	assertDataPacket(t, dummyConn.packetWritten.Front(), 1, []byte{})
	assertDataPacket(t, dummyConn.packetWritten.Back(), 1, []byte{})
}

func TestHandleReadRequest_Timeout(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dummyConn := NewDummyPacketConn("TestHandleReadRequest_ExhaustRetry")

	HandleReadRequest(&dummyConn, []byte{})

	assertDataPacket(t, dummyConn.packetWritten.Front(), 1, []byte{})
	assertErrorPacket(t, dummyConn.packetWritten.Back(), 5, "Send failed")
}

func TestHandleReadRequest_PrematureTermination(t *testing.T) {
	dummyConn := NewDummyPacketConn("TestHandleReadRequest_PrematureTermination",
		packets.NewAck(1),
		packets.NewError(3, "test"))
	file := strings.Repeat("1", MaxPayloadSize)

	HandleReadRequest(&dummyConn, []byte(file))

	assertNumSent(t, dummyConn.packetWritten, 2)
	assertDataPacket(t, dummyConn.packetWritten.Front(), 1, []byte(file))
	assertDataPacket(t, dummyConn.packetWritten.Front().Next(), 2, []byte{})
}

func assertNumSent(t *testing.T, actual *list.List, expected int) {
	t.Helper()
	if actual.Len() != expected {
		t.Error("Incorrect number of packets sent", actual)
	}
}

func assertDataPacket(t *testing.T, actual *list.Element, expectedBlock uint16, expectedData []byte) {
	t.Helper()
	switch actual.Value.(type) {
	default:
		t.Error("Incorrect type packet sent", actual.Value)
	case packets.DataPacket:
		dataPacket := actual.Value.(packets.DataPacket)
		if expectedBlock != dataPacket.Block {
			t.Error("Block incorrect in packet", dataPacket.Block)
		}
		if !bytes.Equal(expectedData, dataPacket.Data) {
			t.Error("Data incorrect in packet", dataPacket.Data)
		}
	}
}

func assertErrorPacket(t *testing.T, actual *list.Element, expectedCode uint16, expectedMsg string) {
	t.Helper()
	switch actual.Value.(type) {
	default:
		t.Error("Incorrect type packet sent", actual.Value)
	case packets.ErrorPacket:
		errorPacket := actual.Value.(packets.ErrorPacket)
		if expectedCode != errorPacket.ErrorCode {
			t.Error("Error code incorrect in packet", errorPacket.ErrorCode)
		}
		if expectedMsg != errorPacket.Message {
			t.Error("Error message incorrect in packet", errorPacket.Message)
		}
	}
}

type DummyPacketConn struct {
	id              string
	packetsToBeRead []packets.Packet
	packetWritten   *list.List
}

func NewDummyPacketConn(id string, packetsToBeRead ...packets.Packet) DummyPacketConn {
	return DummyPacketConn{id: id, packetWritten: list.New(), packetsToBeRead: packetsToBeRead}
}

func (conn *DummyPacketConn) LocalAddr() string {
	return ""
}

func (conn *DummyPacketConn) RemoteAddr() string {
	return conn.id
}

func (conn *DummyPacketConn) Read(timeout time.Duration) (packets.Packet, bool) {
	if len(conn.packetsToBeRead) == 0 {
		return nil, false
	} else {
		top := conn.packetsToBeRead[0]
		conn.packetsToBeRead = conn.packetsToBeRead[1:]
		return top, true
	}
}

func (conn *DummyPacketConn) Write(packet packets.Packet) bool {
	conn.packetWritten.PushBack(packet)
	return true
}

func (conn *DummyPacketConn) Close() {
	/*NO-OP*/
}
