package node

import "github.com/google/uuid"

type Message struct {
	ID        uuid.UUID
	Value     string
	Timestamp int
	Type      MessageType
}

type MessageType int

const (
	Acknowledge MessageType = iota + 1
	OrderGuarantee
)
