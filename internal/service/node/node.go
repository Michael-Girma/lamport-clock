package node

import "fmt"

type Node struct {
	ID    string
	Clock *Clock
}

func (node *Node) SendMessage(msg string) {
	fmt.Printf("Sending message %d", node.Clock.Value)
}

func (node *Node) RecieveMessage() Message {
	return Message{}
}
