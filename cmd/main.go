package main

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	nodegrpc "github.com/michael-girma/lamport-clock/internal/server/grpc"
	"github.com/michael-girma/lamport-clock/internal/service/node"
	"github.com/michael-girma/lamport-clock/internal/utils"
	"google.golang.org/grpc"
)

func main() {
	nodes := make([]*node.Node, 0)

	var wg sync.WaitGroup
	for i := range 15 {
		time.Sleep(1 * time.Second)
		wg.Add(1)
		go func() {
			serviceNode := node.NewNode()
			nodes = append(nodes, serviceNode)
			SetupNode(serviceNode)
			fmt.Printf("Created node %d\n", i)
			wg.Done()
		}()
	}
	PrintNodes(nodes)
	wg.Wait()

	defer func() {
		for _, serviceNode := range nodes {
			serviceNode.Teardown()
		}
	}()
}

func SetupNode(node *node.Node) {
	hostname, port, err := NewAddrForNode()
	if err != nil {
		log.Fatalf("Couldn't create new address for node")
	}

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", *hostname, *port))
	if err != nil {
		log.Fatalf("Couldn't listen on address of new node on port %d: %s", *port, err)
	}

	fmt.Printf("Node %s peers: %d\n", node.ID, len(node.Peers))

	var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(opts...)
	nodegrpc.RegisterNodeServer(grpcServer, node)
	log.Println("Serving GPRC service")

	node.Bootstrap(*hostname, *port)
	time.Sleep(3 * time.Second)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		grpcServer.Serve(lis)
	}()
	wg.Wait()
}

func NewAddrForNode() (*string, *int, error) {
	port, err := utils.GetFreeIPV4Port()
	if err != nil {
		return nil, nil, err
	}
	hostname := "localhost"
	// addr := fmt.Sprintf("localhost:%d", port)
	return &hostname, &port, nil
}

func PrintNodes(nodes []*node.Node) {
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for {
			select {
			case <-ticker.C:
				peerLog := ""
				fmt.Println(len(nodes))
				for _, node := range nodes {
					// nodeObj := *node
					fmt.Printf("Node %s has %d peers\n", node.ID, len(node.Peers))
				}
				fmt.Println(peerLog)
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	wg.Done()
}
