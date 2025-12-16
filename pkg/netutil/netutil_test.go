package netutil

import (
	"net"
	"testing"
)

func TestGetOutboundIP(t *testing.T) {
	ip, err := GetOutboundIP()
	if err != nil {
		t.Fatalf("GetOutboundIP() returned error: %v", err)
	}

	if ip == "" {
		t.Error("GetOutboundIP() returned empty string")
	}

	// Verify it's a valid IP address
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		t.Errorf("GetOutboundIP() returned invalid IP: %s", ip)
	}

	// Verify it's not a loopback address
	if parsedIP.IsLoopback() {
		t.Errorf("GetOutboundIP() returned loopback address: %s", ip)
	}

	// Verify it's an IPv4 address (To4() returns nil for IPv6)
	if parsedIP.To4() == nil {
		t.Logf("Note: GetOutboundIP() returned IPv6 address: %s", ip)
	}

	t.Logf("GetOutboundIP() returned: %s", ip)
}