package online

import (
	"net"
	"testing"
	"time"
)

func TestDiscover(t *testing.T) {
	expected := ServerInfo{Name: "server-1", Port: ":8080"}

	conn, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 0,
	})
	if err != nil {
		t.Fatalf("failed to create server socket: %v", err)
	}
	defer conn.Close()

	discoveryErr := make(chan error, 1)
	go func() {
		discoveryErr <- broadcastReturn(expected, conn)
	}()

	serverAddr := conn.LocalAddr().(*net.UDPAddr)
	clientConn, err := sendDiscoveryRequest(&net.UDPAddr{
		IP:   net.IPv4(127, 0, 0, 1),
		Port: serverAddr.Port,
	})
	if err != nil {
		t.Fatalf("failed to send discovery request to %s: %v", serverAddr, err)
	}
	defer clientConn.Close()

	servers, err := ListenBroadcastReturns(clientConn, time.Second)
	if err != nil {
		t.Fatalf("failed to receive discovery response: %v", err)
	}
	if len(servers) != 1 {
		t.Fatalf("expected 1 server from %s, got %d: %+v", serverAddr, len(servers), servers)
	}

	if servers[0].Name != expected.Name {
		t.Errorf("expected server %q, got %q",
			expected.Name, servers[0].Name)
	}
}
