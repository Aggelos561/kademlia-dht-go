package kademlia

import "sync"

type NodeQueue []*NodeLookup

type PriorityNodes struct {
	pq     NodeQueue
	size   int
	called map[*NodeID]bool
	mu     sync.Mutex
}

// Create a new PriorityNodes instance with a specified size limit
func MakePriorityNodes(size int) *PriorityNodes {
	return &PriorityNodes{
		size:   size,
		called: make(map[*NodeID]bool),
	}
}

// Return the current size of the priority list
func (pn *PriorityNodes) Length() int {
	pn.mu.Lock()
	defer pn.mu.Unlock()

	return len(pn.pq)
}

// Insert an item into the priority list
func (pn *PriorityNodes) Insert(item *NodeLookup) {
	pn.mu.Lock()
	defer pn.mu.Unlock()

	pq := &pn.pq
	insertIndex := len(*pq)
	foundLess := false

	for i, lookup := range *pq {
		if item.Info.Equal(lookup.Info) {
			return
		}
		if item.Less(lookup) {
			if !foundLess {
				foundLess = true
				insertIndex = i
			}
		}
	}

	*pq = append(*pq, nil)
	copy((*pq)[insertIndex+1:], (*pq)[insertIndex:])
	(*pq)[insertIndex] = item

	if len(*pq) > pn.size {
		*pq = (*pq)[:pn.size]
	}
}

// Delete an item from the priority list if it exists.
func (pn *PriorityNodes) Delete(item *NodeLookup) {
	pn.mu.Lock()
	defer pn.mu.Unlock()

	pq := &pn.pq
	for i, lookup := range *pq {
		if item.Info.Equal(lookup.Info) {
			*pq = append((*pq)[:i], (*pq)[i+1:]...)
			return
		}
	}
}

// Merge another PriorityNodes instance into this one
func (pn *PriorityNodes) Merge(pn2 *PriorityNodes) {
	pn2.mu.Lock()
	defer pn2.mu.Unlock()

	pq := pn2.pq
	for _, item := range pq {
		pn.Insert(item)
	}
}

// Set true the called status for a node with the given ID.
func (pn *PriorityNodes) SetCalled(id *NodeID) {
	pn.mu.Lock()
	defer pn.mu.Unlock()

	pn.called[id] = true
}

// Check if a node with the given ID has been called.
func (pn *PriorityNodes) Called(id *NodeID) bool {
	pn.mu.Lock()
	defer pn.mu.Unlock()

	return pn.called[id]
}
