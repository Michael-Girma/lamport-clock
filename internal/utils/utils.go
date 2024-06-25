package utils

import (
	"errors"
	"fmt"
	"net"
	"strconv"
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
	return fmt.Sprintf("LNode-%s", nodeID)
}

func BuildAddressEntry(addressType constants.AddressType, hostname string, port int) string {
	return fmt.Sprintf("%s::%s:%d", addressType, hostname, port)
}

func GetHostAndPortFromEntry(addressEntry string) (*string, *int, error) {
	typeAndAddress := strings.Split(addressEntry, "::")
	if len(typeAndAddress) < 2 {
		return nil, nil, errors.New("invalid address supplied")
	}
	parts := strings.Split(typeAndAddress[1], ":")
	if len(parts) < 2 {
		return nil, nil, errors.New("invalid address supplied")
	}
	port, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, nil, errors.New("invalid port detected in address")
	}
	return &parts[0], &port, nil
}

func GetAddressFromServiceInfo(fields []string, addressType constants.AddressType) *string {
	for _, value := range fields {
		if strings.Contains(value, string(addressType)) {
			parts := strings.Split(value, fmt.Sprintf("%s::", constants.GRPCAddressType))
			if len(parts) < 2 {
				return nil
			}
			return &parts[1]
		}
	}
	return nil
}

func GetAddresssesForNode(entries map[constants.AddressType]string) []string {
	addresses := make([]string, 0)
	for _, value := range entries {
		addresses = append(addresses, value)
	}
	return addresses
}
