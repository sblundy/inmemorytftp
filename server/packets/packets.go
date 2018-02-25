package packets

import (
	"bytes"
	"encoding/binary"
	"io"
)

type OpCode byte

const (
	READ  OpCode = 1
	WRITE OpCode = 2
	DATA  OpCode = 3
	ACK   OpCode = 4
	ERROR OpCode = 5
)

type Packet interface {
	Bytes() []byte
}

type ReadPacket struct {
	Filename string
	Mode     string
}

type WritePacket struct {
	Filename string
	Mode     string
}

type DataPacket struct {
	Block uint16
	Data  []byte
}

type AckPacket struct {
	Block uint16
}

type ErrorPacket struct {
	ErrorCode uint16
	Message   string
}

func (packet ReadPacket) Bytes() []byte {
	panic("Not supported")
}

func (packet WritePacket) Bytes() []byte {
	panic("Not supported")
}

func (packet DataPacket) Bytes() []byte {
	buff := newPacketBuffer(DATA)
	writeUint16ToBuff(buff, packet.Block)
	buff.Write(packet.Data)
	return buff.Bytes()
}

func (packet AckPacket) Bytes() []byte {
	buff := newPacketBuffer(ACK)
	writeUint16ToBuff(buff, packet.Block)
	return buff.Bytes()
}

func (packet ErrorPacket) Bytes() []byte {
	buff := newPacketBuffer(ERROR)
	writeUint16ToBuff(buff, packet.ErrorCode)
	buff.WriteString(packet.Message)
	buff.WriteByte(0)
	return buff.Bytes()
}

func newPacketBuffer(code OpCode) *bytes.Buffer {
	buff := bytes.NewBuffer(make([]byte, 0))
	buff.WriteByte(0)
	buff.WriteByte(byte(code))
	return buff
}

func writeUint16ToBuff(buffer io.Writer, i uint16) (int, error) {
	err := binary.Write(buffer, binary.BigEndian, i)
	if err != nil {
		return -1, err
	}
	return 2, err
}

func bytesToInt16(bytes []byte) uint16 {
	return binary.BigEndian.Uint16(bytes[:2])
}

func Read(bytes []byte) (Packet, bool) {
	if len(bytes) < 2 {
		return nil, false
	} else if bytes[0] != 0 {
		return nil, false
	} else if OpCode(bytes[1]) < READ || ERROR < OpCode(bytes[1]) {
		return nil, false
	}
	switch OpCode(bytes[1]) {
	default:
		panic("Should be unreachable")
	case READ:
		filename, n := readPacketString(bytes[2:])
		mode, _ := readPacketString(bytes[2+n:])
		return ReadPacket{Filename: filename, Mode: mode}, true
	case WRITE:
		filename, n := readPacketString(bytes[2:])
		mode, _ := readPacketString(bytes[2+n:])
		return WritePacket{Filename: filename, Mode: mode}, true
	case DATA:
		block := bytesToInt16(bytes[2:4])
		return DataPacket{Block: block, Data: bytes[4:]}, true
	case ACK:
		block := bytesToInt16(bytes[2:4])
		return AckPacket{Block: block}, true
	case ERROR:
		errorCode := bytesToInt16(bytes[2:4])
		msg, _ := readPacketString(bytes[4:])
		return ErrorPacket{ErrorCode: errorCode, Message: msg}, true
	}
}

func readPacketString(payload []byte) (string, int) {
	n := bytes.IndexByte(payload, 0)
	return string(payload[:n]), n + 1
}

func NewData(block uint16, data []byte) DataPacket {
	return DataPacket{Block: block, Data: data}
}

func NewAck(block uint16) AckPacket {
	return AckPacket{Block: block}
}

func NewError(errorCode uint16, msg string) ErrorPacket {
	return ErrorPacket{ErrorCode: errorCode, Message: msg}
}
