package node

import (
	"fmt"

	"github.com/google/uuid"
)

type Node struct {
	ID               uuid.UUID
	Clock            *Clock
	MessageQueue     []*Message
	Acknowledgements map[uuid.UUID][]uuid.UUID
}

func NewNode() *Node {
	return &Node{
		ID:               uuid.New(),
		Clock:            NewClock(),
		MessageQueue:     make([]*Message, 0),
		Acknowledgements: make(map[uuid.UUID][]uuid.UUID),
	}
}

func (node *Node) SendMessage(msgValue string) {
	node.Clock.Mutex.Lock()
	defer node.Clock.Mutex.Unlock()
	message := &Message{
		ID:        uuid.New(),
		Timestamp: node.Clock.Value,
		Value:     msgValue,
	}
	node.MessageQueue = append(node.MessageQueue, message)
	node.Acknowledgements[message.ID] = make([]uuid.UUID, 0)
	fmt.Printf("Sending message %d", node.Clock.Value)
}

func (node *Node) RecieveMessage() Message {
	return Message{}
}

func (node *Node) Boostrap() {
	// Use this function to bootstrap a node within a network of distributed systems
}
