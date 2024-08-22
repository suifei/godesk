package network

import (
    "encoding/binary"
    "io"
    "net"

    "github.com/suifei/godesk/internal/protocol"
    "google.golang.org/protobuf/proto"
)

type TCPConnection struct {
    conn net.Conn
}

func NewTCPConnection(conn net.Conn) *TCPConnection {
    return &TCPConnection{conn: conn}
}

func (c *TCPConnection) Send(msg *protocol.Message) error {
    data, err := proto.Marshal(msg)
    if err != nil {
        return err
    }

    // 先发送消息长度
    lengthBuf := make([]byte, 4)
    binary.BigEndian.PutUint32(lengthBuf, uint32(len(data)))
    _, err = c.conn.Write(lengthBuf)
    if err != nil {
        return err
    }

    // 再发送消息内容
    _, err = c.conn.Write(data)
    return err
}

func (c *TCPConnection) Receive() (*protocol.Message, error) {
    // 先读取消息长度
    lengthBuf := make([]byte, 4)
    _, err := io.ReadFull(c.conn, lengthBuf)
    if err != nil {
        return nil, err
    }
    length := binary.BigEndian.Uint32(lengthBuf)

    // 再读取消息内容
    data := make([]byte, length)
    _, err = io.ReadFull(c.conn, data)
    if err != nil {
        return nil, err
    }

    // 解析消息
    msg := &protocol.Message{}
    err = proto.Unmarshal(data, msg)
    if err != nil {
        return nil, err
    }

    return msg, nil
}

func (c *TCPConnection) Close() error {
    return c.conn.Close()
}