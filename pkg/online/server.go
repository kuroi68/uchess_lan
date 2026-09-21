package online

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net"
)

type ServerInfo struct {
	Name string `json:"name"`
	Port string `json:"port"`
}

func sendServerInfo(server ServerInfo, conn *net.UDPConn, remoteAddr *net.UDPAddr) error {
	data, err := json.Marshal(server)
	if err != nil {
		return fmt.Errorf("json marshal error: %v", err)
	}
	// отправляем ответ клиенту
	_, err = conn.WriteToUDP(data, remoteAddr)
	if err != nil {
		return fmt.Errorf("send response error: %v", err)
	}

	return nil
}

func broadcastReturn(server ServerInfo, conn *net.UDPConn) error {
	buffer := make([]byte, 1024)

	for {
		n, remoteAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}

			var netErr net.Error
			if errors.As(err, &netErr) && netErr.Timeout() {
				// Timeout можно повторить.
				continue
			}

			// Неизвестная или постоянная ошибка сокета.
			return fmt.Errorf("read discovery request: %w", err)
		}

		if string(buffer[:n]) != discoverMessage {
			continue
		}

		log.Printf("Получен запрос от %s\n", remoteAddr)

		if err := sendServerInfo(server, conn, remoteAddr); err != nil {
			log.Printf("Не удалось отправить инфу о сервере")
			continue
		}
	}
}

func Host(port, name string) (*Session, error) {
	listenerTCP, err := net.Listen("tcp", port)
	if err != nil {
		return nil, fmt.Errorf("Listen TCP: %w", err)
	}
	defer listenerTCP.Close()

	udpConn, err := net.ListenUDP("udp", &net.UDPAddr{
		Port: discoverPort,
		IP:   net.ParseIP("0.0.0.0"), // Слушаем на всех интерфейсах
	})
	if err != nil {
		return nil, fmt.Errorf("Listen UDP discovery: %w", err)
	}
	defer udpConn.Close()

	server := ServerInfo{
		Name: name,
		Port: port,
	}

	discoveryErr := make(chan error, 1)
	go func() {
		discoveryErr <- broadcastReturn(server, udpConn)
	}()

	type acceptResult struct {
		conn net.Conn
		err  error
	}

	acceptCh := make(chan acceptResult, 1)
	go func() {
		conn, err := listenerTCP.Accept()
		acceptCh <- acceptResult{conn: conn, err: err}
	}()

	select {
	case err := <-discoveryErr:
		if err != nil {
			return nil, fmt.Errorf("discovery failed: %w", err)
		}
		return nil, errors.New("discovery stopped unexpectedly")

	case result := <-acceptCh:
		// Клиент подключился — discovery больше не нужен.
		_ = udpConn.Close()

		if result.err != nil {
			return nil, fmt.Errorf("accept TCP connection: %w", result.err)
		}

		return newSession(result.conn), nil
	}
}
