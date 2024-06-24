package node

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/mdns"
	"github.com/michael-girma/lamport-clock/internal/constants"
	nodegrpc "github.com/michael-girma/lamport-clock/internal/server/grpc"
	"github.com/michael-girma/lamport-clock/internal/utils"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Node struct {
	nodegrpc.UnimplementedNodeServer
	ID               uuid.UUID
	Clock            *Clock
	MessageQueue     []*nodegrpc.Message
	Acknowledgements map[string][]string
	Peers            []*Peer
	MDNSServer       *mdns.Server
	ServiceAddresses map[constants.AddressType]string
}

type Peer struct {
	ServiceEntry *mdns.ServiceEntry
	Client       nodegrpc.NodeClient
}

func NewNode() *Node {

	return &Node{
		ID:               uuid.New(),
		Clock:            NewClock(),
		MessageQueue:     make([]*nodegrpc.Message, 0),
		Acknowledgements: make(map[string][]string),
		ServiceAddresses: make(map[constants.AddressType]string),
	}
}

func (node *Node) SendMessage(msgValue string) {
	node.Clock.Mutex.Lock()
	defer node.Clock.Mutex.Unlock()
	message := &nodegrpc.Message{
		ID:        uuid.New().String(),
		Timestamp: node.Clock.Value,
		Value:     msgValue,
	}
	node.MessageQueue = append(node.MessageQueue, message)
	node.Acknowledgements[message.ID] = make([]string, 0)
	fmt.Printf("Sending message %d", node.Clock.Value)
}

func (node Node) RecieveMessage(ctx context.Context, msg *nodegrpc.Message) (*nodegrpc.Message, error) {
	fmt.Printf("RecieveMessage called on node %s\n", node.ID)
	return &nodegrpc.Message{}, nil
}

func (node *Node) AdvertiseServiceOnMDNS() {
	// Setup our service export
	hostname, _ := os.Hostname()
	info := []string{
		fmt.Sprintf("clock server node-%s", node.ID),
		fmt.Sprintf("%s:%s", constants.NodeID, node.ID),
		node.ServiceAddresses[constants.GRPCAddressType],
		node.ServiceAddresses[constants.MDNSAddressType],
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

	node.ServiceAddresses[constants.MDNSAddressType] = utils.BuildAddressEntry(constants.MDNSAddressType, hostname, port)
	node.MDNSServer = server
}

func (node *Node) Bootstrap(hostname string, port int) {
	node.ServiceAddresses[constants.GRPCAddressType] = utils.BuildAddressEntry(constants.GRPCAddressType, hostname, port)
	node.AdvertiseServiceOnMDNS()
	time.Sleep(3 * time.Second)
	node.LookupExistingNodes()
}

func (node *Node) LookupExistingNodes() {
	// Make a channel for results and start listening
	entriesCh := make(chan *mdns.ServiceEntry, 10) // Change entry size to use env var
	go func() {
		for entry := range entriesCh {
			nodeID := utils.GetNodeIDFromServiceInfo(entry.InfoFields)
			if nodeID != nil && *nodeID != node.ID.String() {
				grpcClient, err := GenerateNodeClient(entry)
				if err != nil {
					log.Printf("Couldn't create client for peer %s: %s\n", *nodeID, err)
					continue
				}
				peer := &Peer{
					ServiceEntry: entry,
					Client:       *grpcClient,
				}
				node.Peers = append(node.Peers, peer)
				fmt.Printf("Got new entry: %s\n", *nodeID)
			}
		}
	}()

	// Start the lookup
	err := mdns.Lookup(string(constants.LamportService), entriesCh)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Closing mdns entries channel. Discovered %d peers \n", len(node.Peers))
	// close(entriesCh)
}

func GenerateNodeClient(entry *mdns.ServiceEntry) (*nodegrpc.NodeClient, error) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))

	grpcAddressOfClient := utils.GetAddressFromServiceInfo(entry.InfoFields, constants.GRPCAddressType)
	if grpcAddressOfClient == nil {
		return nil, errors.New("couldn't generate grpc client for peer")
	}

	conn, err := grpc.NewClient(fmt.Sprintf("%s:%s", "http://", *grpcAddressOfClient), opts...)
	if err != nil {
		var nodeID = utils.GetNodeIDFromServiceInfo(entry.InfoFields)
		return nil, fmt.Errorf("couldn't create grpc client for peer %s at address %s: %s", *nodeID, *grpcAddressOfClient, err)
	}
	client := nodegrpc.NewNodeClient(conn)
	return &client, nil
}

func (node *Node) Teardown() {
	if node.MDNSServer != nil {
		node.MDNSServer.Shutdown()
	}

	// Send leave message to peers
}
