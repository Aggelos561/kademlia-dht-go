package kademlia

import (
	"container/list"
	"errors"
	"math/big"
	"sync"
	"time"

	"github.com/Aggelos561/kademlia-dht-go/logging"
)

const BucketLength = 160
const K = 10

type KBuckets struct {
	mutex        sync.Mutex
	RoutingTable *Tree
}

// Initialize a new KBuckets struct with an empty tree-based routing table
func MakeBuckets(id *NodeID) *KBuckets {
	b := new(KBuckets)
	b.RoutingTable = MakeTree(id)
	return b
}

// Insert a node into the appropriate K-bucket based on the XOR-based distance
func (b *KBuckets) PushBucket(key BitArray, info *NodeInfo) *NodeInfo {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.RoutingTable.Insert(key, info)
}

// Return a list of the K closest nodes to the given key
func (b *KBuckets) FindNodes(key BitArray) *list.List {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.RoutingTable.Closest(key)
}

// Get index of K-bucket (in flat representation) given two node IDs
// Used only to calculate expiration times
func FindIndex(id1 *NodeID, id2 *NodeID) (uint, error) {
	distance := CalcDist(id1, id2)

	var i uint = 1
	n := big.NewInt(2)

	for ; i <= BucketLength; i++ {
		m := big.NewInt(int64(i))
		num := new(big.Int).Exp(n, m, nil)

		if num.Cmp(distance) > 0 {
			return i - 1, nil
		}
	}
	return 0, errors.New("range does not exist")
}

// Refreshes the K-Buckets by doing a lookup on the last node in each bucket and pinging all nodes in the returned list
func (node *Node) RefreshBuckets(hours int64, a int) {
	for {
		time.Sleep(time.Duration(hours) * time.Hour)

		node.Buckets.mutex.Lock()
		currentTime := time.Now().Unix()
		elapsedTime := hours * 3600

		buckets, LastUpdates := node.Buckets.RoutingTable.AllBuckets()
		for i, bucket := range buckets { // loop should be parallelized in the future
			LastUpdate := LastUpdates[i]
			if currentTime > (LastUpdate + elapsedTime) { // refresh this bucket
				n := bucket.Back()
				if n == nil {
					continue
				}
				pq, _, err := node.Lookup(n.Value.(*NodeInfo).ID, uint(a))
				if err != nil {
					logging.Warning("%s", err)
					continue
				} else {
					logging.Info("refreshing bucket %d", i)
				}
				for _, item := range pq {
					err := node.CallPing(item.Info)
					if err != nil {
						logging.Warning("%s", err)
					}
				}
			}
		}
		node.Buckets.mutex.Unlock()
	}
}
