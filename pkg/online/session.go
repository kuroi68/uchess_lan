package online

import (
	"bufio"
	"encoding/binary"
	"io"
	"net"
	"sync"
)

type Session struct {
	conn net.Conn
	mu   sync.Mutex
	r    *bufio.Reader
}

func newSession(conn net.Conn) *Session {
	return &Session{conn: conn, r: bufio.NewReader(conn)}
}

func (p *Session) Send(data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var length [4]byte
	binary.BigEndian.PutUint32(length[:], uint32(len(data)))
	if _, err := p.conn.Write(length[:]); err != nil {
		return err
	}

	_, err := p.conn.Write(data)
	return err
}

func (p *Session) Receive() ([]byte, error) {
	var length [4]byte
	if _, err := io.ReadFull(p.r, length[:]); err != nil {
		return nil, err
	}

	n := binary.BigEndian.Uint32(length[:])
	buf := make([]byte, n)
	if _, err := io.ReadFull(p.r, buf); err != nil {
		return nil, err
	}

	return buf, nil
}

func (p *Session) Close() error {
	return p.conn.Close()
}
