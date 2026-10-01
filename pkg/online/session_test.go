// session_test.go
package online

import (
	"io"
	"net"
	"testing"
	"time"
)

func TestSessionSendAndClose(t *testing.T) {
	clientConn, serverConn := net.Pipe()

	clientSession := newSession(clientConn)
	serverSession := newSession(serverConn)
	defer serverSession.Close()

	received := make(chan []byte, 1)
	receiveErr := make(chan error, 1)
	go func() {
		buf, err := serverSession.Receive()
		if err != nil {
			receiveErr <- err
			return
		}
		received <- buf
	}()

	if err := clientSession.Send([]byte("hello")); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	select {
	case err := <-receiveErr:
		t.Fatalf("Receive() error = %v", err)
	case got := <-received:
		if string(got) != "hello" {
			t.Fatalf("got %q, want %q", got, "hello")
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for data")
	}

	if err := clientSession.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// После Close Send должен возвращать ошибку.
	if err := clientSession.Send([]byte("x")); err == nil {
		t.Fatal("Send() after Close() = nil error, want error")
	}
}

func TestSessionSendConcurrentSafe(t *testing.T) {
	clientConn, serverConn := net.Pipe()
	defer clientConn.Close()
	defer serverConn.Close()

	sess := newSession(clientConn)

	go func() {
		buf := make([]byte, 2)
		for {
			if _, err := io.ReadFull(serverConn, buf); err != nil {
				return
			}
		}
	}()

	const n = 20
	errs := make(chan error, n)
	for i := 0; i < n; i++ {
		go func() { errs <- sess.Send([]byte("ab")) }()
	}
	for i := 0; i < n; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("Send() error = %v", err)
		}
	}
}
