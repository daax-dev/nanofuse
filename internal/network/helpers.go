package network

import (
	"crypto/rand"
	"fmt"
)

// GenerateMAC generates a random MAC address with Firecracker OUI prefix
// Format: AA:FC:00:xx:xx:xx
// AA:FC:00 is the Firecracker OUI (Organizationally Unique Identifier)
func GenerateMAC() (string, error) {
	// Generate 3 random bytes for the last part of MAC
	buf := make([]byte, 3)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("failed to generate random MAC: %w", err)
	}

	// Format as MAC address with Firecracker OUI
	mac := fmt.Sprintf("AA:FC:00:%02X:%02X:%02X",
		buf[0], buf[1], buf[2])

	return mac, nil
}

// ValidateMAC checks if a MAC address is valid
// Basic validation: checks format XX:XX:XX:XX:XX:XX
func ValidateMAC(mac string) bool {
	if len(mac) != 17 {
		return false
	}

	for i, c := range mac {
		if i%3 == 2 {
			// Every third character should be ':'
			if c != ':' {
				return false
			}
		} else {
			// Other characters should be hex digits
			if !((c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')) {
				return false
			}
		}
	}

	return true
}
