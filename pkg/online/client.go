package online

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"time"
)

const (
	discoverMessage = "UCHESS_DISCOVER v1"
	discoverPort    = 8081
)

func sendDiscoveryRequest(addr *net.UDPAddr) (*net.UDPConn, error) {
	// Создаём UDP-сокет на случайном локальном порту.
	// Именно на этот порт серверы будут отвечать.
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{
		IP:   net.IPv4zero,
		Port: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("listen udp: %w", err)
	}

	message := []byte(discoverMessage)
	_, err = conn.WriteToUDP(message, addr)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("send broadcast: %w", err)
	}
	return conn, nil
}

func SendBroadcast() (*net.UDPConn, error) {
	return sendDiscoveryRequest(&net.UDPAddr{
		IP:   net.IPv4bcast, // 255.255.255.255
		Port: discoverPort,
	})
}

func ListenBroadcastReturns(conn *net.UDPConn, timeout time.Duration) ([]ServerInfo, error) {
	return listenBroadcastReturns(conn, timeout, false)
}

func listenBroadcastReturns(conn *net.UDPConn, timeout time.Duration, firstOnly bool) ([]ServerInfo, error) {
	err := conn.SetReadDeadline(time.Now().Add(timeout))
	if err != nil {
		return nil, fmt.Errorf("set deadline: %w", err)
	}
	var servers []ServerInfo
	buffer := make([]byte, 1024)
	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			var netErr net.Error
			// Timeout здесь означает:
			// "время поиска серверов закончилось".
			if errors.As(err, &netErr) && netErr.Timeout() {
				return servers, nil
			}
			return nil, fmt.Errorf("read udp: %w", err)
		}

		var server ServerInfo
		if err := json.Unmarshal(buffer[:n], &server); err != nil {
			continue
		}

		host, port, err := net.SplitHostPort(server.Port)
		if err != nil {
			continue
		}

		parsedIP := net.ParseIP(host)
		if host == "" || parsedIP != nil && parsedIP.IsUnspecified() {
			host = remoteAddr.IP.String()
		}

		server.Port = net.JoinHostPort(host, port)
		servers = append(servers, server)
		if firstOnly {
			return servers, nil
		}
	}
}

func Discover(timeout time.Duration) ([]ServerInfo, error) {
	return discover(timeout, false)
}

func DiscoverFirst(timeout time.Duration) ([]ServerInfo, error) {
	return discover(timeout, true)
}

func discover(timeout time.Duration, firstOnly bool) ([]ServerInfo, error) {
	conn, err := SendBroadcast()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	servers, err := listenBroadcastReturns(conn, timeout, firstOnly)
	if err != nil {
		return nil, err
	}
	return servers, nil
}

// TODO: реализовать возможность пользователю выбрать хост из Discover
func Connect(addr string) (*Session, error) {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return newSession(conn), nil
}
