package kademlia

import (
	"math/big"
)

type NodeLookup struct {
	Info *NodeInfo
	Dist big.Int
}

// Create a new NodeLookup instance with the given NodeInfo and distance
func MakeNodeLookup(info *NodeInfo, dist big.Int) *NodeLookup {
	return &NodeLookup{
		Info: info,
		Dist: dist,
	}
}

// Compare two NodeLookup instances based on their distance
func (nd *NodeLookup) Less(n *NodeLookup) bool {
	return nd.Dist.Cmp(&n.Dist) < 0
}
