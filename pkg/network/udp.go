package network

import (
	"net"

	"github.com/suifei/godesk/internal/protocol"
	"google.golang.org/protobuf/proto"
)

type UDPConnection struct {
	conn *net.UDPConn
	addr *net.UDPAddr
}

func NewUDPConnection(conn *net.UDPConn, addr *net.UDPAddr) *UDPConnection {
	return &UDPConnection{conn: conn, addr: addr}
}

func (c *UDPConnection) Send(msg *protocol.Message) error {
	data, err := proto.Marshal(msg)
	if err != nil {
		return err
	}

	_, err = c.conn.WriteToUDP(data, c.addr)
	return err
}

func (c *UDPConnection) Receive() (*protocol.Message, error) {
	buffer := make([]byte, 65507) // UDP的最大包大小
	n, _, err := c.conn.ReadFromUDP(buffer)
	if err != nil {
		return nil, err
	}

	msg := &protocol.Message{}
	err = proto.Unmarshal(buffer[:n], msg)
	if err != nil {
		return nil, err
	}

	return msg, nil
}

func (c *UDPConnection) Close() error {
	return c.conn.Close()
}
