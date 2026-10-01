// client_test.go (дополнение к тому, что уже есть в online_test.go)
package online

import (
	"net"
	"testing"
	"time"
)

func TestConnectSendsOverRealTCP(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start test listener: %v", err)
	}
	defer ln.Close()

	received := make(chan []byte, 1)
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		serverSession := newSession(conn)
		buf, _ := serverSession.Receive()
		received <- buf
	}()

	sess, err := Connect(ln.Addr().String())
	if err != nil {
		t.Fatalf("Connect() error = %v", err)
	}
	defer sess.Close()

	if err := sess.Send([]byte("ping")); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	select {
	case got := <-received:
		if string(got) != "ping" {
			t.Fatalf("got %q, want %q", got, "ping")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for server to receive data")
	}
}
