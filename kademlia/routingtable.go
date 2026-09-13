package kademlia

import (
	"container/list"
	"time"
)

type Tree struct {
	owner  *NodeID
	root   *BTNode
	depth  int
	leaves int
}

type BTNode struct {
	left       *BTNode
	right      *BTNode
	bucket     *list.List // Each K-bucket is a list
	lastUpdate int64      // Unix timestamp for last update to K-bucket
	level      int
}

// Returns a new Tree.
func MakeTree(id *NodeID) *Tree {
	tr := &Tree{}
	tr.root = MakeBTNode(0)
	tr.root.left = MakeBTNode(1)
	tr.root.right = MakeBTNode(1)
	tr.root.bucket = nil
	tr.depth = 1
	tr.leaves = 2
	tr.owner = id
	return tr
}

// Returns a new BTNode.
func MakeBTNode(level int) *BTNode {
	nd := &BTNode{left: nil, right: nil, bucket: list.New(), lastUpdate: time.Now().Unix(), level: level}
	return nd
}

// Copy source list to destination list
func copyList(src *list.List, dest *list.List) {
	for e := src.Front(); e != nil && dest.Len() != K; e = e.Next() {
		node := e.Value.(*NodeInfo)
		dest.PushBack(node)
	}
}

// Insert a key to the tree-structured routing table.
func (tr *Tree) Insert(key BitArray, info *NodeInfo) *NodeInfo {
	return tr.root.insert(tr, key, info)
}

// Attempt to insert a key to the tree. Full k-buckets containing the node's ID are split
// Relaxed splitting is supported to deal with highly unbalanced trees
func (nd *BTNode) insert(tr *Tree, key BitArray, info *NodeInfo) *NodeInfo {
	if nd.bucket != nil {
		if nd.bucket.Len() == K && ((key.bits[nd.level-1] == 0) || (key.bits[nd.level-1] == 1 && (nd.level%B != 0))) && nd.level != BucketLength {
			nd.left = MakeBTNode(nd.level + 1)
			nd.right = MakeBTNode(nd.level + 1)
			tr.depth = max(tr.depth, nd.level+1)
			tr.leaves++ // remember why this and not +2

			for e := nd.bucket.Front(); e != nil; e = e.Next() {
				node := e.Value.(*NodeInfo)
				dist := XorBits(tr.owner, node.ID)
				if dist.bits[nd.level] == 1 {
					nd.left.bucket.PushBack(node)
				} else {
					nd.right.bucket.PushBack(node)
				}
			}
			nd.bucket = nil
		} else {
			return nd.UpdateBucket(info)
		}
	}
	if key.bits[nd.level] == 1 {
		nd.left.insert(tr, key, info)
	} else {
		nd.right.insert(tr, key, info)
	}
	return nil
}

// Update the K-bucket with the given node.
// If the K-bucket is full, the oldest node is removed and returned.
// If the node already exists in the K-bucket, it is moved to the back of the list.
// Otherwise, the node is inserted in the K-bucket.
func (nd *BTNode) UpdateBucket(info *NodeInfo) *NodeInfo {
	bucket := nd.bucket
	for e := bucket.Front(); e != nil; e = e.Next() {
		node := e.Value.(*NodeInfo)
		if node.Equal(info) {
			bucket.MoveToBack(e)
			nd.lastUpdate = time.Now().Unix()
			return nil
		}
	}
	if bucket.Len() != K {
		bucket.PushBack(info)
		nd.lastUpdate = time.Now().Unix()
	} else { // bucket is full
		node := bucket.Remove(bucket.Front()).(*NodeInfo)
		return node
	}
	return nil
}

// Get k-closest nodes to the given key
// If the first k-bucket is not full extend search to neighboring subtrees to fill up to k
func (tr *Tree) Closest(key BitArray) *list.List {
	kclosest := list.New()
	tr.root.closest(tr, key, kclosest)
	return kclosest
}

func (nd *BTNode) closest(tr *Tree, key BitArray, kclosest *list.List) {
	var origin int

	if nd.bucket != nil {
		copyList(nd.bucket, kclosest)
		return
	}

	if key.bits[nd.level] == 1 {
		nd.left.closest(tr, key, kclosest)
		origin = 1 // left child origin
	} else {
		nd.right.closest(tr, key, kclosest)
		origin = 0 // right child origin
	}

	if kclosest.Len() == K {
		return
	} else if origin == 1 {
		nd.right.closest(tr, key, kclosest)
	} else {
		nd.left.closest(tr, key, kclosest)
	}
}

// Get list of all k-buckets of routing table in order to support bucket refresh operation
// The last update timestamp of each bucket is also returned as a separate list
func (tr *Tree) AllBuckets() ([]*list.List, []int64) {
	buckets := make([]*list.List, 0)
	timestamps := make([]int64, 0)
	tr.root.allBuckets(tr, &buckets, &timestamps)
	return buckets, timestamps
}

func (nd *BTNode) allBuckets(tr *Tree, buckets *[]*list.List, timestamps *[]int64) {
	if nd.bucket != nil {
		*buckets = append(*buckets, nd.bucket)
		*timestamps = append(*timestamps, nd.lastUpdate)
		return
	}
	nd.right.allBuckets(tr, buckets, timestamps)
	nd.left.allBuckets(tr, buckets, timestamps)
}
