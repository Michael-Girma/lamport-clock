package utils

import (
	"fmt"
	"net"
	"strings"

	"github.com/google/uuid"
	"github.com/michael-girma/lamport-clock/internal/constants"
)

func GetFreeIPV4Port() (int, error) {
	// Create a listener on port 0 to get a free port
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()

	// Get the assigned address and extract the port number
	addr := l.Addr().(*net.TCPAddr)
	return addr.Port, nil
}

func GetNodeIDFromServiceInfo(fields []string) *string {
	for _, value := range fields {
		if strings.HasPrefix(value, string(constants.NodeID)) {
			potentialValue := strings.Split(value, ":")[1]
			if potentialValue != "" {
				return &potentialValue
			}
		}
	}
	return nil
}

func GenerateNodeInstanceName(nodeID uuid.UUID) string {
	return fmt.Sprint("LNode-%s", nodeID)
}
