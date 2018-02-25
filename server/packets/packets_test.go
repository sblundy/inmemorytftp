package packets

import (
	"bytes"
	"testing"
)

func TestRead_ReadEmptyPacket(t *testing.T) {
	_, ok := Read([]byte{})

	if ok {
		t.Error("Read of empty packet should have failed")
	}
}

func TestRead_ReadInvalidOpcode1(t *testing.T) {
	_, ok := Read([]byte{1, 0})

	if ok {
		t.Error("Read of empty packet should have failed")
	}
}

func TestRead_ReadInvalidOpcode2(t *testing.T) {
	_, ok := Read([]byte{0, 0})

	if ok {
		t.Error("Read of empty packet should have failed")
	}
}

func TestRead_ReadPacket(t *testing.T) {
	bytesBuilder := bytes.NewBuffer([]byte{0, byte(READ)})
	writePacketString(bytesBuilder, "test.txt")
	writePacketString(bytesBuilder, "binary")

	output, ok := Read(bytesBuilder.Bytes())

	if !ok {
		t.Error("Read failed")
	} else {
		switch output.(type) {
		default:
			t.Error("Wrong type", output)
		case ReadPacket:
			packet := output.(ReadPacket)
			if "test.txt" != packet.Filename {
				t.Error("Filename incorrect", packet.Filename)
			}
			if "binary" != packet.Mode {
				t.Error("Mode incorrect", packet.Mode)
			}
		}
	}
}

func TestRead_WritePacket(t *testing.T) {
	bytesBuilder := bytes.NewBuffer([]byte{0, byte(WRITE)})
	writePacketString(bytesBuilder, "test.txt")
	writePacketString(bytesBuilder, "binary")

	output, ok := Read(bytesBuilder.Bytes())

	if !ok {
		t.Error("Read failed")
	} else {
		switch output.(type) {
		default:
			t.Error("Wrong type", output)
		case WritePacket:
			packet := output.(WritePacket)
			if "test.txt" != packet.Filename {
				t.Error("Filename incorrect", packet.Filename)
			}
			if "binary" != packet.Mode {
				t.Error("Mode incorrect", packet.Mode)
			}
		}
	}
}

func TestRead_DataPacket(t *testing.T) {
	bytesBuilder := bytes.NewBuffer([]byte{0, byte(DATA), 0, 1})
	bytesBuilder.WriteString("payload")

	output, ok := Read(bytesBuilder.Bytes())

	if !ok {
		t.Error("Read failed")
	} else {
		switch output.(type) {
		default:
			t.Error("Wrong type", output)
		case DataPacket:
			packet := output.(DataPacket)
			if 1 != packet.Block {
				t.Error("Block incorrect", packet.Block)
			}
			if !bytes.Equal([]byte("payload"), packet.Data) {
				t.Error("Data incorrect", packet.Data)
			}
		}
	}
}

func TestRead_AckPacket(t *testing.T) {
	bytesBuilder := bytes.NewBuffer([]byte{0, byte(ACK), 0, 1})

	output, ok := Read(bytesBuilder.Bytes())

	if !ok {
		t.Error("Read failed")
	} else {
		switch output.(type) {
		default:
			t.Error("Wrong type", output)
		case AckPacket:
			packet := output.(AckPacket)
			if 1 != packet.Block {
				t.Error("Block incorrect", packet.Block)
			}
		}
	}
}

func TestRead_ErrorPacket(t *testing.T) {
	bytesBuilder := bytes.NewBuffer([]byte{0, byte(ERROR), 0, 1})
	writePacketString(bytesBuilder, "error message")

	output, ok := Read(bytesBuilder.Bytes())

	if !ok {
		t.Error("Read failed")
	} else {
		switch output.(type) {
		default:
			t.Error("Wrong type", output)
		case ErrorPacket:
			packet := output.(ErrorPacket)
			if 1 != packet.ErrorCode {
				t.Error("ErrorCode incorrect", packet.ErrorCode)
			}
			if "error message" != packet.Message {
				t.Error("Message incorrect", packet.Message)
			}
		}
	}
}

func writePacketString(buff *bytes.Buffer, s string) {
	buff.WriteString(s)
	buff.WriteByte(0)
}

func TestDataPacket_Bytes(t *testing.T) {
	sut := NewData(1, []byte("payload"))
	output := sut.Bytes()

	assert2ByteCodeEqual(output[:2], 0, byte(DATA), t, "opcode incorrect")
	assert2ByteCodeEqual(output[2:4], 0, 1, t, "Block incorrect")

	if !bytes.Equal(output[4:], []byte("payload")) {
		t.Error("payload incorrect", output[4:])
	}
	if len(output) > 2+2+7 {
		t.Error("Extranious bytes", output[11:])
	}
}

func TestAckPacket_Bytes(t *testing.T) {
	sut := NewAck(1)
	output := sut.Bytes()

	assert2ByteCodeEqual(output[:2], 0, byte(ACK), t, "opcode incorrect")
	assert2ByteCodeEqual(output[2:4], 0, 1, t, "Block incorrect")
	if len(output) > 2+2 {
		t.Error("Extranious bytes", output[4:])
	}
}

func TestErrorPacket_Bytes(t *testing.T) {
	sut := NewError(1, "error message")
	output := sut.Bytes()

	assert2ByteCodeEqual(output[:2], 0, byte(ERROR), t, "opcode incorrect")
	assert2ByteCodeEqual(output[2:4], 0, 1, t, "error code incorrect")

	if len(output) != 2+2+13+1 {
		t.Error("Length incorrect", len(output))
	}
	if !bytes.Equal(output[4:17], []byte("error message")) {
		t.Error("error message incorrect", output[4:2+2+len("error message")])
	}
	if output[17] != 0 {
		t.Error("Invalid terminator", output[17])
	}
}

func assert2ByteCodeEqual(bytes []byte, b1 byte, b2 byte, t *testing.T, msg string) {
	if !(len(bytes) == 2 && bytes[0] == b1 && bytes[1] == b2) {
		t.Error(msg, bytes)
	}
}
