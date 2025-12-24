package netutil

import (
	"errors"
	"net"
)

var ErrNoValidIP = errors.New("no valid IP address found")

// GetOutboundIP returns the preferred outbound IP address of this machine.
// It uses a UDP dial to determine which local IP would be used to reach
// an external destination, without actually sending any data.
func GetOutboundIP() (string, error) {
	// Use UDP dial to determine outbound IP (no data sent)
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}