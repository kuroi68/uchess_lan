package online

import (
	"net"
	"sync"
)

type Session struct {
	conn net.Conn
	mu   sync.Mutex
}

func (p *Session) Send(data []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	_, err := p.conn.Write(data)
	return err
}

func (p *Session) Close() error {
	return p.conn.Close()
}

func newSession(conn net.Conn) *Session {
	return &Session{conn: conn}
}
