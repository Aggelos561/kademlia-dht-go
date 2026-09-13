package kademlia

import (
	"fmt"
	"slices"
	"testing"
)

// Testing Tree creation and initialization
func TestTree_Make(t *testing.T) {
	tr := MakeTree(nil)
	if tr == nil {
		t.Fatal("MakeTree() constructor returned nil")
	}
	if tr.root == nil {
		t.Fatal("Tree root is nil")
	}
	if tr.depth != 1 {
		t.Errorf("Expected tree depth 1, got %d", tr.depth)
	}
	if tr.leaves != 2 {
		t.Errorf("Expected tree leaves 2, got %d", tr.leaves)
	}
}

// Testing Tree insertion
func TestTree_Insert(t *testing.T) {
	owner := BitArrayToBytes(*MakeBitArray(BucketLength))
	tr := MakeTree((*NodeID)(&owner))
	insertedNodes := make([]*NodeInfo, 0)
	insertedKeys := make([]*BitArray, 0)

	for i := 1; i <= 10; i++ {
		info := MakeNode("127.0.0.1", fmt.Sprint(4000+i)).Info
		key := MakeBitArray(BucketLength)
		key.bits[BucketLength-i] = 1
		tr.Insert(*key, info)
		insertedNodes = append(insertedNodes, info)
		insertedKeys = append(insertedKeys, key)
	}

	buckets, _ := tr.AllBuckets()
	for idx, node := range insertedNodes {
		found := false
		for _, bucket := range buckets {
			for e := bucket.Front(); e != nil; e = e.Next() {
				bucketNode := e.Value.(*NodeInfo)
				if bucketNode.Equal(node) {
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			t.Errorf("Inserted node with key %v not found in any bucket", insertedKeys[idx].bits)
		}
	}
}

// Testing closest nodes to a given key
func TestTree_Closest(t *testing.T) {
	owner := BitArrayToBytes(*MakeBitArray(BucketLength))
	tr := MakeTree((*NodeID)(&owner))
	var inserted []*NodeInfo

	for i := 1; i <= (K * 2); i++ {
		info := MakeNode("127.0.0.1", fmt.Sprint(5000+i)).Info
		key := MakeBitArray(BucketLength)
		key.bits[BucketLength-i] = 1
		tr.Insert(*key, info)
		inserted = append(inserted, info)
	}

	// Searching for a key with the last bit (LSB) set to 1
	searchKey := MakeBitArray(BucketLength)
	searchKey.bits[BucketLength-1] = 1
	closest := tr.Closest(*searchKey)

	if closest.Len() > K {
		t.Errorf("Closest returned more than K nodes: %d", closest.Len())
	}

	for e := closest.Front(); e != nil; e = e.Next() {
		node := e.Value.(*NodeInfo)
		found := slices.ContainsFunc(inserted, node.Equal)
		if !found {
			t.Errorf("node %v in closest not in inserted nodes", node.ID)
		}
	}
}
