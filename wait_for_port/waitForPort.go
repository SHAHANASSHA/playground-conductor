package wait_for_port

import (
	"fmt"
	"net"
	"time"
)

func WaitForPort(host string, port string, timeout time.Duration, interval time.Duration) error {
	address := net.JoinHostPort(host, port)
	start := time.Now()

	for {
		conn, err := net.DialTimeout("tcp", address, 2*time.Second)
		if err == nil {
			conn.Close()
			return nil
		}

		if time.Since(start) > timeout {
			return fmt.Errorf("timeout waiting for port %s", address)
		}

		time.Sleep(interval)
	}
}
