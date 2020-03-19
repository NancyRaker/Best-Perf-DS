package hashring

import (
	"testing"
)

const (
	node1 = "192.168.0.1"
	node2 = "192.168.0.2"
	node3 = "192.168.0.3"
)

func getNodesCount(nodes nodesArray) (int, int, int) {
	node1Count := 0
	node2Count := 0
	node3Count := 0

	for _, node := range nodes {
		if node.key == node1 {
			node1Count += 1
		}
		if node.key == node2 {
			node2Count += 1

		}
		if node.key == node3 {
			node3Count += 1
		}
	}
	return node1Count, node2Count, node3Count
}


func TestNewHashRing(t *testing.T) {
	hr := NewHashRing(100)
	hr.AddNodes(map[string]int{
		node1: 4,
		node2: 2,
		node3: 2,
	})

	i, i2, i3 := getNodesCount(hr.nodes)
	t.Logf("node1 count: %d, node2 count: %d, node3 count: %d, ", i, i2, i3 )

	value := hr.GetNode("127.0.0.1")
	t.Logf("127.0.0.1 get ip : %s", value)

	hr.UpdateNode(node2, 4)
	i, i2, i3 = getNodesCount(hr.nodes)
	t.Logf("after update node2, node1 count: %d, node2 count: %d, node3 count: %d, ", i, i2, i3 )

	value = hr.GetNode("127.0.0.1")
	t.Logf("127.0.0.1 get ip : %s", value)

	hr.RemoveNodes([]string{node2})
	i, i2, i3 = getNodesCount(hr.nodes)
	t.Logf("after remove node2, node1 count: %d, node2 count: %d, node3 count: %d, ", i, i2, i3 )

	value = hr.GetNode("127.0.0.1")
	t.Logf("127.0.0.1 get ip : %s", value)
}
