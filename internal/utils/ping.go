package utils

import (
	"net"
	"strings"
	"time"
)

type PingStatus string

const (
	PingUp   PingStatus = "UP"
	PingDown PingStatus = "DOWN"
)

func PingHost(addr string) PingStatus {
	conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return PingDown
	}
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	buf := make([]byte, 256)
	n, err := conn.Read(buf)
	if err != nil {
		return PingDown
	}

	if strings.Contains(string(buf[:n]), "SSH-") {
		return PingUp
	}

	return PingDown
}
