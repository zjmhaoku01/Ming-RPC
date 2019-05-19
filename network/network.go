package network

import (
	"encoding/binary"
	"io"
	"net"
)

// Transport struct
type Transport struct {
	conn *net.Conn
}

// New Transport
func New(conn *net.Conn) *Transport {
	return &Transport{conn}
}

// Send data
// 数据格式：数据长度（4字节） + 数据
func (t *Transport) Send(data *[]byte) error {
	buf := make([]byte, 4+len(*data))
	binary.BigEndian.PutUint32(buf[:4], uint32(len(*data))) // Set Header field
	copy(buf[4:], *data)                                    // Set Data field
	_, err := (*t.conn).Write(buf)
	return err
}

// Receive data
// 先读数据长度（4字节）再读数据
func (t *Transport) Receive() ([]byte, error) {
	header := make([]byte, 4)
	_, err := io.ReadFull(*t.conn, header)
	if err != nil {
		return nil, err
	}
	dataLen := binary.BigEndian.Uint32(header)
	data := make([]byte, dataLen)
	_, err = io.ReadFull(*t.conn, data)
	if err != nil {
		return nil, err
	}
	return data, err
}
