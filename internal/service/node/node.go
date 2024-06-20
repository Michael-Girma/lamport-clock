package node

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/mdns"
	"github.com/michael-girma/lamport-clock/internal/constants"
	"github.com/michael-girma/lamport-clock/internal/utils"
)

type Node struct {
	ID               uuid.UUID
	Clock            *Clock
	MessageQueue     []*Message
	Acknowledgements map[uuid.UUID][]uuid.UUID
	Peers            []*mdns.ServiceEntry
	MDNSServer       *mdns.Server
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
	node.setupMDNSServer()
	time.Sleep(3 * time.Second)
	node.lookupExistingNodes()
}

func (node *Node) setupMDNSServer() {
	// Setup our service export
	// host, _ := os.Hostname()
	info := []string{
		fmt.Sprintf("clock server node-%s", node.ID),
		fmt.Sprintf("%s:%s", constants.NodeID, node.ID),
	}
	port, err := utils.GetFreeIPV4Port()
	if err != nil {
		panic("Couldn't bootstrap node: No available port")
	}
	serviceInstanceName := utils.GenerateNodeInstanceName(node.ID)
	service, _ := mdns.NewMDNSService(serviceInstanceName, string(constants.LamportService), "", "", port, nil, info)

	// Create the mDNS server, defer shutdown
	fmt.Printf("Creating mdns server on port: %d for node: %s \n", port, node.ID)
	server, err := mdns.NewServer(&mdns.Config{Zone: service})

	if err != nil {
		panic(fmt.Sprintf("Error creating mdns server %s", err))
	}

	node.MDNSServer = server
}

func (node *Node) lookupExistingNodes() {
	// Make a channel for results and start listening
	entriesCh := make(chan *mdns.ServiceEntry, 10) // Change entry size to use env var
	go func() {
		for entry := range entriesCh {
			nodeID := utils.GetNodeIDFromServiceInfo(entry.InfoFields)
			if nodeID != nil && *nodeID != node.ID.String() {
				node.Peers = append(node.Peers, entry)
				fmt.Printf("Got new entry: %s\n", *nodeID)
			}
		}
	}()

	// Start the lookup
	err := mdns.Lookup(string(constants.LamportService), entriesCh)
	if err != nil {
		panic(err)
	}
	fmt.Println("Closing mdns entries channel")
	// close(entriesCh)
}

func (node *Node) Teardown() {
	if node.MDNSServer != nil {
		node.MDNSServer.Shutdown()
	}

	// Send leave message to peers
}
