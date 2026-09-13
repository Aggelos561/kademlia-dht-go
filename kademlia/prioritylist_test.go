package kademlia

import (
	"fmt"
	"math/big"
	"testing"
)

// Testing priority list ascending order based on distance
func TestPriorityNodes_InsertOrdersByDistance(t *testing.T) {
	pn := MakePriorityNodes(K)

	for i := int64(K); i >= 1; i-- {
		pn.Insert(MakeNodeLookup(MakeNode("127.0.0.1", fmt.Sprint(4000+i)).Info, *big.NewInt(i)))
	}

	for i := 1; i < pn.Length(); i++ {
		if !(pn.pq)[i-1].Less((pn.pq)[i]) {
			t.Errorf("priority list not sorted: item %d is not < item %d", i-1, i)
		}
	}
}

// Testing priority list duplicate insertions
func TestPriorityNodes_InsertAvoidsDuplicates(t *testing.T) {
	pn := MakePriorityNodes(K)

	lookup := MakeNodeLookup(MakeNode("127.0.0.1", "4000").Info, *big.NewInt(10))

	pn.Insert(lookup)
	pn.Insert(lookup)

	if pn.Length() != 1 {
		t.Errorf("expected 1 unique item, got %d", pn.Length())
	}
}

// Testing priority list K limit size
func TestPriorityNodes_EnforcesLimit(t *testing.T) {
	pn := MakePriorityNodes(K)

	for i := int64(1); i <= (K * 2); i++ {
		pn.Insert(MakeNodeLookup(MakeNode("127.0.0.1", fmt.Sprint(4000+i)).Info, *big.NewInt(i)))
	}

	if pn.Length() != K {
		t.Errorf("priority list exceeded max size: got %d, want 20", pn.Length())
	}
}

// Testing priority list Merging functionality
func TestPriorityNodes_Merge(t *testing.T) {
	pn1 := MakePriorityNodes(K)
	pn2 := MakePriorityNodes(K)

	for i := int64(1); i <= K/2; i++ {
		pn1.Insert(MakeNodeLookup(MakeNode("127.0.0.1", fmt.Sprint(4000+i)).Info, *big.NewInt(i)))
		pn2.Insert(MakeNodeLookup(MakeNode("127.0.0.1", fmt.Sprint(4100+i)).Info, *big.NewInt(i + 10)))
	}

	// Merging the two priority lists
	pn1.Merge(pn2)

	if pn1.Length() != K {
		t.Errorf("merge should result in 20 items, got %d", pn1.Length())
	}

	pq1 := pn1.pq
	for i := 1; i < len(pq1); i++ {
		if pq1[i-1].Dist.Cmp(&pq1[i].Dist) > 0 {
			t.Errorf("priority list not sorted: item %d (dist %s) is not <= item %d (dist %s)", i-1, pq1[i-1].Dist.String(), i, pq1[i].Dist.String())
		}
	}
}

// Testing priority list Delete functionality
func TestPriorityNodes_Delete(t *testing.T) {
	pn := MakePriorityNodes(K)

	for i := int64(K); i >= 1; i-- {
		pn.Insert(MakeNodeLookup(MakeNode("127.0.0.1", fmt.Sprint(4000+i)).Info, *big.NewInt(i)))
	}
	pn.Delete(pn.pq[2])

	pq := pn.pq

	if len(pq) == K {
		t.Errorf("deletion error")
	}

	for i := 1; i < len(pq); i++ {
		if !(pq)[i-1].Less((pq)[i]) {
			t.Errorf("priority list not sorted: item %d is not < item %d", i-1, i)
		}
	}
}

func TestPriorityNodes_Called(t *testing.T) {
	pn := MakePriorityNodes(K)

	nodes := make([]*Node, 0)

	for i := int64(1); i <= K; i++ {
		node := MakeNode("127.0.0.1", fmt.Sprint(4000+i))
		nodes = append(nodes, node)

		pn.Insert(MakeNodeLookup(node.Info, *big.NewInt(i)))
	}

	calledNode := nodes[0].Info.ID
	pn.SetCalled(calledNode)

	if !pn.called[calledNode] {
		t.Errorf("updating called failed")
	}
}

// Testing priority list Size with one item
func TestPriorityNodes_Size_One(t *testing.T) {
	pn := MakePriorityNodes(1)

	portIndex := K / 2
	for range K / 2 {
		pn.Insert(MakeNodeLookup(MakeNode("127.0.0.1", fmt.Sprint(4000+portIndex)).Info, *big.NewInt(int64(portIndex))))
		portIndex++
	}

	portIndex = 0
	for range K / 2 {
		pn.Insert(MakeNodeLookup(MakeNode("127.0.0.1", fmt.Sprint(4000+portIndex)).Info, *big.NewInt(int64(portIndex))))
		portIndex++
	}

	if pn.Length() != 1 {
		t.Errorf("error inserting nodes, expected 1 node, got %d", pn.Length())
	}

	if pn.pq[0].Dist.Cmp(big.NewInt(0)) != 0 {
		t.Errorf("error comparing nodes, expected distance 0, got %s", pn.pq[0].Dist.String())
	}

}
