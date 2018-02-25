package connection

import (
	"github.com/sblundy/inmemorytftp/server/packets"
	"log"
	"net"
	"time"
)

type TftpPacketConn interface {
	LocalAddr() string
	RemoteAddr() string
	Read(timeout time.Duration) (packets.Packet, bool)
	Write(packet packets.Packet) bool
	Close()
}

type TftpReplyChannel interface {
	Write(packet packets.Packet) bool
}

type Connection struct {
	conn  *net.UDPConn
	raddr net.Addr
}

type ResponseChannel struct {
	conn  net.PacketConn
	raddr net.Addr
}

func New(destination net.Addr) (TftpPacketConn, error) {
	laddr, err := net.ResolveUDPAddr("udp", ":")
	if err != nil {
		return nil, err
	}

	conn, err := net.ListenUDP("udp", laddr)

	return &Connection{conn: conn, raddr: destination}, nil
}

func WrapExisting(conn net.PacketConn, raddr net.Addr) TftpReplyChannel {
	return &ResponseChannel{conn: conn, raddr: raddr}
}

func (conn *Connection) LocalAddr() string {
	return conn.conn.LocalAddr().String()
}

func (conn *Connection) RemoteAddr() string {
	return conn.raddr.String()
}

func (conn *Connection) Read(timeout time.Duration) (packets.Packet, bool) {
	buff := make([]byte, 516)
	conn.conn.SetReadDeadline(time.Now().Add(timeout))
	n, err := conn.conn.Read(buff)
	if err != nil {
		log.Println("Error reading", err)
		return nil, false
	}
	return packets.Read(buff[:n])
}

func (conn *Connection) Write(packet packets.Packet) bool {
	_, err := conn.conn.WriteTo(packet.Bytes(), conn.raddr)
	if err != nil {
		log.Println("Error writing", err)
		return false
	}
	return true
}

func (conn *Connection) Close() {
	conn.conn.Close()
}

func (conn *ResponseChannel) Write(packet packets.Packet) bool {
	_, err := conn.conn.WriteTo(packet.Bytes(), conn.raddr)
	if err != nil {
		log.Println("Error writing", err)
		return false
	}
	return true
}
