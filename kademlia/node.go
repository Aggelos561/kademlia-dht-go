package kademlia

import (
	"bytes"
	"container/list"
	"crypto/rand"
	"database/sql"
	"fmt"
	"log"
	"math/big"
	"net"
	"net/http"
	"net/rpc"
	"runtime"
	"sync"
	"time"

	"github.com/Aggelos561/kademlia-dht-go/logging"
)

const Timeout = 10000 * time.Millisecond
const IDLength = 20
const CACHING = 1
const B = 5

type NodeID [IDLength]byte

// Hold the information about a node in the Kademlia network.
type NodeInfo struct {
	ID   *NodeID
	IP   string
	Port string
}

// Hold the information about a node itself in the Kademlia network, along with its K-bucket and key-value database.
// Used by a running Kademlia node.
type Node struct {
	Info    *NodeInfo
	Buckets *KBuckets
	DB      *sql.DB
	dbMux   sync.Mutex
}

// Check if two NodeInfo objects are equal based on their IDs.
func (nf *NodeInfo) Equal(info *NodeInfo) bool {
	return bytes.Equal(nf.ID[:], info.ID[:])
}

// Generate a random NodeID of length IDLength.
func generateID() (*NodeID, error) {
	id := new(NodeID)
	_, err := rand.Read(id[:])

	if err != nil {
		return nil, err
	}
	return id, nil
}

// Create a new node with a randomly generated ID, the given IP and port, and initialize its K-buckets and key-value database.
// For testing purposes only.
func MakeNode(ip string, port string) *Node {
	id, err := generateID()
	if err != nil {
		log.Fatal(err)
	}
	buckets := MakeBuckets(id)

	db := MakeDB(*id)

	nodeInfo := &NodeInfo{
		ID:   id,
		IP:   ip,
		Port: port,
	}

	return &Node{
		Info:    nodeInfo,
		Buckets: buckets,
		DB:      db,
	}
}

// Create a new node with the given ID, IP and port, and initialize its K-buckets and key-value database.
func MakeNodeStatic(id *NodeID, ip string, port string) *Node {
	buckets := MakeBuckets(id)
	db := MakeDB(*id)

	nodeInfo := &NodeInfo{
		ID:   id,
		IP:   ip,
		Port: port,
	}

	return &Node{
		Info:    nodeInfo,
		Buckets: buckets,
		DB:      db,
	}
}

// Listening starts the RPC server for the given node.
func (node *Node) Listening() {
	rpc.Register(node)
	rpc.HandleHTTP()

	port := ":" + node.Info.Port
	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal("listen error: ", err)
	}

	http.Serve(l, nil)
}

// Establish a TCP connection to another node using its NodeInfo.
func (node *Node) EstablishConnection(info *NodeInfo) (*rpc.Client, error) {
	client, err := rpc.DialHTTP("tcp", info.IP+":"+info.Port)
	if err != nil {
		return nil, fmt.Errorf("establishing TCP connection timeout [%s:%s → %s:%s]", node.Info.IP, node.Info.Port, info.IP, info.Port)
	}

	return client, nil
}

// Find the K closest nodes to a given NodeID in the node's K-buckets.
func (node *Node) FindClosest(id *NodeID) *list.List {
	dist := XorBits(node.Info.ID, id)
	kclosest := node.Buckets.FindNodes(dist)
	return kclosest
}

// Insert nodes into a priority queue based on their distance from the given id
func (node *Node) populatePn(pn *PriorityNodes, items any, id *NodeID) {
	switch v := items.(type) {

	case *list.List:
		for e := v.Front(); e != nil; e = e.Next() {
			info := e.Value.(*NodeInfo)
			if node.Info.Equal(info) {
				continue
			}
			dist := CalcDist(id, info.ID)
			pn.Insert(MakeNodeLookup(info, *dist))
		}

	case []*NodeInfo:
		for _, info := range v {
			if node.Info.Equal(info) {
				continue
			}
			dist := CalcDist(id, info.ID)
			pn.Insert(MakeNodeLookup(info, *dist))
		}
	}
}

// Expanding lookup when closest node does not change in node lookup operations
func (node *Node) expandLookup(pn *PriorityNodes, start int, unresponsive *PriorityNodes, id *NodeID) *PriorityNodes {
	var wg sync.WaitGroup
	rem := MakePriorityNodes(K)
	totalThreads := runtime.NumCPU()

	if len(pn.pq) < (start + 1) {
		return rem
	}

	pq := pn.pq[start:]
	iters := int(len(pq) / totalThreads)

	wg.Add(totalThreads)
	for i := range totalThreads {
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			begin := i * iters
			end := begin + iters
			if i == totalThreads-1 {
				end = len(pq)
			}

			for _, item := range pq[begin:end] {
				if !pn.Called(item.Info.ID) {
					kclosest, err := node.CallFindNode(item.Info)
					if err == nil {
						node.populatePn(rem, kclosest, id)
						pn.SetCalled(item.Info.ID)
						if unresponsive != nil {
							unresponsive.Delete(item)
						}
					} else {
						if unresponsive != nil {
							unresponsive.Insert(item)
						}
					}
				}
			}
		}(&wg)
	}
	wg.Wait()

	return rem
}

// Expanding lookup when closest node does not change in value lookup operations
func (node *Node) expandValueLookup(pn *PriorityNodes, start int, id *NodeID, unresponsive *PriorityNodes, cachingPn *PriorityNodes) (string, *PriorityNodes) {
	var wg sync.WaitGroup
	var mu sync.Mutex
	rem := MakePriorityNodes(K)
	value := ""

	totalThreads := runtime.NumCPU()

	if len(pn.pq) < (start + 1) {
		return value, rem
	}

	pq := pn.pq[start:]
	iters := int(len(pq) / totalThreads)

	wg.Add(totalThreads)
	for i := range totalThreads {
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			begin := i * iters
			end := begin + iters
			if i == totalThreads-1 {
				end = len(pq)
			}

			for _, item := range pq[begin:end] {
				if !pn.Called(item.Info.ID) {
					ret, err := node.CallFindValue(item.Info, id)
					kclosest, localValue := ret.Nodes, ret.Value
					if localValue != "" {
						mu.Lock()
						value = localValue
						mu.Unlock()
						return
					}
					if err == nil {
						node.populatePn(rem, kclosest, id)
						pn.SetCalled(item.Info.ID)
						cachingPn.Insert(item)
						if unresponsive != nil {
							unresponsive.Delete(item)
						}
					} else {
						if unresponsive != nil {
							unresponsive.Insert(item)
						}
					}
				}
			}
		}(&wg)
	}
	wg.Wait()

	return value, rem
}

// NodeLookup recursive algorithm (Kademlia paper)
func (node *Node) Lookup(id *NodeID, a uint) (NodeQueue, uint, error) {
	var wg sync.WaitGroup
	var iter uint

	pn := MakePriorityNodes(K)
	unresponsive := MakePriorityNodes(K)

	node.populatePn(pn, node.FindClosest(id), id)
	if pn.Length() == 0 {
		return NodeQueue{}, iter, fmt.Errorf("buckets are empty")
	}

	aMax := min(pn.Length(), int(a))
	logging.Info("node lookup for key %x starting\n", *id)
	for {
		iter++
		closest := pn.pq[0]
		closest_a := pn.pq[0:aMax]
		for index, item := range closest_a {
			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				logging.Info("node lookup thread %d for key %x\n", index, *id)
				defer wg.Done()
				threadPn := MakePriorityNodes(K)
				kclosest, err := node.CallFindNode(item.Info)
				if err != nil {
					unresponsive.Insert(item)
				} else {
					node.populatePn(threadPn, kclosest, id)
					pn.Merge(threadPn)
					pn.SetCalled(item.Info.ID)
				}
			}(&wg)
		}
		wg.Wait()

		if closest.Info.Equal(pn.pq[0].Info) {
			expandedPn := node.expandLookup(pn, aMax, unresponsive, id)
			pn.Merge(expandedPn)

			if closest.Info.Equal(pn.pq[0].Info) {
				expandedPn = node.expandLookup(unresponsive, 0, nil, id)
				pn.Merge(expandedPn)

				if closest.Info.Equal(pn.pq[0].Info) { // there is nothing else to do
					break
				}
			}
		}
	}

	logging.Info("node lookup for key %x took: %d iters\n", *id, iter)
	return pn.pq, iter, nil
}

// Cache the value in the closest node to the key.
// Used in ValueLookup to cache the value in the closest node, not containing the key, after a successful lookup
func (node *Node) cacheStore(cachingPn *PriorityNodes, key *NodeID, value *string) {
	if *value == "" || cachingPn.Length() == 0 {
		return
	}
	for _, item := range cachingPn.pq {
		err := node.CallStore(item.Info, key, *value, true)
		if err != nil {
			logging.Warning("%s", err)
		} else {
			logging.Info("caching key %x with value \"%s\" successfully [%s:%s → %s:%s]\n", *key, *value, node.Info.IP, node.Info.Port, item.Info.IP, item.Info.Port)
		}
	}
}

// ValueLookup recursive algorithm (Kademlia paper)
func (node *Node) ValueLookup(id *NodeID, a uint) (string, uint, error) {
	var mu sync.Mutex
	var wg sync.WaitGroup
	var iter uint
	var value string = ""

	pn := MakePriorityNodes(K)
	unresponsive := MakePriorityNodes(K)
	cachingPN := MakePriorityNodes(CACHING)
	defer node.cacheStore(cachingPN, id, &value)

	node.populatePn(pn, node.FindClosest(id), id)
	if pn.Length() == 0 {
		return "", iter, fmt.Errorf("buckets are empty")
	}

	aMax := min(pn.Length(), int(a))
	logging.Info("value lookup for key %x starting\n", *id)
	for {
		iter++
		closest := pn.pq[0]
		closest_a := pn.pq[0:aMax]
		for index, item := range closest_a {
			wg.Add(1)
			go func(value *string, wg *sync.WaitGroup) {
				logging.Info("value lookup thread %d for key %x\n", index, *id)
				defer wg.Done()
				threadPn := MakePriorityNodes(K)
				ret, err := node.CallFindValue(item.Info, id)
				kclosest, val := ret.Nodes, ret.Value
				if val != "" {
					mu.Lock()
					*value = val
					mu.Unlock()
					return
				}
				if err != nil {
					unresponsive.Insert(item)
				} else {
					node.populatePn(threadPn, kclosest, id)
					pn.Merge(threadPn)
					pn.SetCalled(item.Info.ID)
					cachingPN.Insert(item)
				}
			}(&value, &wg)
		}
		wg.Wait()

		if value != "" {
			logging.Info("value lookup for key %x took: %d iters\n", *id, iter)
			return value, iter, nil
		}

		if closest.Info.Equal(pn.pq[0].Info) {
			value, expandedPn := node.expandValueLookup(pn, aMax, id, unresponsive, cachingPN)
			if value != "" {
				logging.Info("value lookup for key %x took: %d iters (after expanded lookup)\n", *id, iter)
				return value, iter, nil
			}
			pn.Merge(expandedPn)

			if closest.Info.Equal(pn.pq[0].Info) {
				value, expandedPn = node.expandValueLookup(unresponsive, 0, id, nil, cachingPN)
				if value != "" {
					logging.Info("value lookup for key %x took: %d iters (after unresponsive lookup)\n", *id, iter)
					return value, iter, nil
				}
				pn.Merge(expandedPn)

				if closest.Info.Equal(pn.pq[0].Info) { // there is nothing else to do
					break
				}
			}
		}
	}

	return "", iter, fmt.Errorf("value lookup failed after %d iters for key: %x", iter, *id)
}

// Update the K-bucket for a given node by calculating the distance and updating the K-bucket accordingly.
func (node *Node) PingHandler(info *NodeInfo) {
	dist := XorBits(node.Info.ID, info.ID)
	old_info := node.Buckets.PushBucket(dist, info)
	if old_info != nil { // K-bucket is full
		err := node.CallPing(old_info)
		if err != nil { // Ping to old node was not successfull
			node.Buckets.PushBucket(dist, info)
		}
	}
}

// Handle the RPC call and wait for a reply or timeout.
func rpcHandler(rpcType string, divCall *rpc.Call, node *Node, info *NodeInfo, client *rpc.Client) error {
	defer client.Close()
	select {
	case replyCall := <-divCall.Done:
		if replyCall.Error != nil {
			return fmt.Errorf("[caller] %s RPC error [%s:%s → %s:%s]: %s", rpcType, node.Info.IP, node.Info.Port, info.IP, info.Port, replyCall.Error)
		} else {
			node.PingHandler(info)
			logging.Info("[caller] %s RPC success [%s:%s → %s:%s]\n", rpcType, node.Info.IP, node.Info.Port, info.IP, info.Port)
		}
		return nil

	case <-time.After(Timeout):
		return fmt.Errorf("[caller] %s RPC timeout [%s:%s → %s:%s]", rpcType, node.Info.IP, node.Info.Port, info.IP, info.Port)
	}
}

// Send a Ping RPC to the specified node.
func (node *Node) CallPing(info *NodeInfo) error {
	client, err := node.EstablishConnection(info)
	if err != nil {
		return err
	}

	rpcType := "Node.Ping"
	args := &PingArgs{Sender: node.Info}
	var reply any

	divCall := client.Go(rpcType, args, &reply, nil)

	return rpcHandler(rpcType, divCall, node, info, client)
}

// Send a FindNode RPC to the specified node and return the k-closest nodes.
func (node *Node) CallFindNode(info *NodeInfo) ([]*NodeInfo, error) {
	client, err := node.EstablishConnection(info)
	if err != nil {
		return nil, err
	}

	rpcType := "Node.FindNode"
	args := &FindNodeArgs{Sender: node.Info, ID: info.ID}
	reply := make([]*NodeInfo, 0)

	divCall := client.Go(rpcType, args, &reply, nil)

	err = rpcHandler(rpcType, divCall, node, info, client)

	return reply, err
}

// Send a Store RPC to the specified node to store a key-value pair.
func (node *Node) CallStore(info *NodeInfo, key *NodeID, value string, caching bool) error {
	client, err := node.EstablishConnection(info)
	if err != nil {
		return err
	}

	rpcType := "Node.Store"
	args := &StoreArgs{Sender: node.Info, Key: key, Value: value, Caching: caching}
	var reply any

	divCall := client.Go(rpcType, args, &reply, nil)

	return rpcHandler(rpcType, divCall, node, info, client)
}

// Send a FindValue RPC to the specified node to retrieve the value associated with a given key.
func (node *Node) CallFindValue(info *NodeInfo, key *NodeID) (FindValueRet, error) {
	client, err := node.EstablishConnection(info)
	if err != nil {
		return FindValueRet{}, err
	}

	rpcType := "Node.FindValue"
	args := &FindValueArgs{Sender: node.Info, ID: key}
	reply := FindValueRet{}

	divCall := client.Go(rpcType, args, &reply, nil)

	err = rpcHandler(rpcType, divCall, node, info, client)

	return reply, err
}

// Bootstrapping a node by pinging a list of known nodes and then performing a lookup for the node's own ID
func (node *Node) Bootstrap(nodes []*NodeInfo, a int) {
	for _, item := range nodes {
		err := node.CallPing(item)
		if err != nil {
			logging.Warning("%s", err)
		}
	}

	randInt, _ := rand.Int(rand.Reader, big.NewInt(3))
	time.Sleep(time.Duration(randInt.Int64()) * time.Second)

	myClosest, _, err := node.Lookup(node.Info.ID, uint(a))
	if err != nil {
		logging.Error("%s", err)
		return
	}

	for _, item := range myClosest {
		err := node.CallPing(item.Info)
		if err != nil {
			logging.Warning("%s", err)
		}
	}
}
