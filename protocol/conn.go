package protocol

import (
	"encoding/json"
	"log"
	"net"

	"golem/protocol/protocol"
)

// BytesReader is the interface that wraps the ReadBytes method,
// similar to ReadByte in io.ByteReader.
type BytesReader interface {
	ReadBytes(int) ([]byte, error)
}

// A ClientConn implements io.Reader, io.ByteReader, BytesReader,
// io.Writer, and io.Closer by wrapping a net.Conn.
type ClientConn struct {
	conn   net.Conn
	logger *log.Logger
}

// NewClientConn returns a new ClientConn from a net.Conn
// and an optional packet logger for debugging.
func NewClientConn(conn net.Conn, logger *log.Logger) *ClientConn {
	c := ClientConn{}
	c.conn = conn
	c.logger = logger
	return &c
}

// ReadHandshakePacket reads a handshake packet.
func (c ClientConn) ReadHandshakePacket() (protocol.HandshakePacket, error) {
	var p protocol.HandshakePacket
	err := c.readPacket(&p, protocol.HandshakePacketID)
	return p, err
}

// ReadStatusRequestPacket reads a status request packet.
func (c ClientConn) ReadStatusRequestPacket() (protocol.StatusRequestPacket, error) {
	var p protocol.StatusRequestPacket
	err := c.readPacket(&p, protocol.StatusRequestPacketID)
	return p, err
}

// ReadLoginStartPacket reads a login start packet.
func (c ClientConn) ReadLoginStartPacket() (protocol.LoginStartPacket, error) {
	var p protocol.LoginStartPacket
	err := c.readPacket(&p, protocol.LoginStartPacketID)
	return p, err
}

// ReadAndRespondPing reads a ping packet and sends a pong.
func (c *ClientConn) ReadAndRespondPing() error {

	_, data, err := readPacket(c, protocol.StatusPingPacketID)
	if err != nil {
		return err
	}

	_, err = c.Write(data)
	return err

}

// WriteMessageStatus sends a status message.
func (c *ClientConn) WriteMessageStatus(
	text string,
	versionName string,
	versionProtocol int,
	playersOnline int,
	playersMax int,
) error {

	serverStatus := protocol.ServerStatus{
		protocol.Version{
			versionName,
			versionProtocol,
		},
		protocol.Players{
			playersMax,
			playersOnline,
			[]protocol.PlayerSample{},
		},
		protocol.Description{text},
		"",
	}

	bytes, err := json.Marshal(serverStatus)
	if err != nil {
		return err
	}

	p := protocol.StatusResponsePacket{string(bytes)}
	return c.writePacket(&p, protocol.StatusResponsePacketID)

}

// WriteMessageText sends a text message.
func (c *ClientConn) WriteMessageText(text string) error {

	serverText := protocol.ServerText{text}

	bytes, err := json.Marshal(serverText)
	if err != nil {
		return err
	}

	p := protocol.StatusResponsePacket{string(bytes)}
	return c.writePacket(&p, protocol.StatusResponsePacketID)

}

// Read implements the io.Reader interface.
func (c *ClientConn) Read(p []byte) (int, error) {
	n, err := c.conn.Read(p)
	if c.logger != nil {
		c.logger.Printf("read: %x\n", p[:n])
	}
	return n, err
}

// ReadByte implements the io.ByteReader interface.
func (c *ClientConn) ReadByte() (byte, error) {
	b, err := c.ReadBytes(1)
	return b[0], err
}

// ReadN implements the BytesReader interface.
func (c *ClientConn) ReadBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := c.Read(b)
	return b, err
}

// Write implements the io.Writer interface.
func (c *ClientConn) Write(p []byte) (int, error) {
	n, err := c.conn.Write(p)
	if c.logger != nil {
		c.logger.Printf("write: %x\n", p)
	}
	return n, err
}

// Close implements the io.Closer interface.
func (c *ClientConn) Close() error {
	return c.conn.Close()
}
