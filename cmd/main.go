package main

import (
	"fmt"

	"github.com/michael-girma/lamport-clock/internal/service/node"
)

func main() {
	nodes := make([]node.Node, 0)
	for i := range 10 {
		serviceNode := node.NewNode()
		serviceNode.Boostrap()
		fmt.Println("Created node %d", i)
		nodes = append(nodes, *serviceNode)
	}

	fmt.Printf("Total Nodes: %d\n", len(nodes))

	for i, serviceNode := range nodes {
		fmt.Printf("Node %d peers: %d\n", i, len(serviceNode.Peers))
		serviceNode.Teardown()
	}
}
