package node

type INode interface {
	SendMessage(msg string)
	RecieveMessage() Message
	ListenForPeers()
}
