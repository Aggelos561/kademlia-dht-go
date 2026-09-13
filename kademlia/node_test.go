package kademlia

import (
	"crypto/sha1"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"testing"
	"time"
)

func deleteDbs() {
	err := os.RemoveAll("./dbs")
	if err != nil {
		fmt.Printf("error removing dbs directory: %s\n", err)
	}
}

// Deleting sqlite files, registering HTTP global handler & disabling logs
func TestMain(m *testing.M) {
	deleteDbs()

	rpc.HandleHTTP()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	exitCode := m.Run()
	deleteDbs()

	os.Exit(exitCode)
}

// Testing basic Kademlia operations
func TestNode_Operations(t *testing.T) {
	// Node1 will be the client
	var ID1 NodeID
	var ID2 NodeID
	for i := range 20 {
		ID2[i] = 0xFF
	}

	node1 := MakeNodeStatic(&ID1, "127.0.0.1", "4000")
	node2 := MakeNodeStatic(&ID2, "127.0.0.1", "4001")

	// Node2 starts to listen (server node)
	go func(n *Node) {
		srv := rpc.NewServer()
		srv.Register(n)

		mux := http.NewServeMux()
		mux.Handle("/_goRPC_", srv)

		port := ":" + n.Info.Port
		l, err := net.Listen("tcp", port)
		if err != nil {
			t.Errorf("tcp listening error: %s", err)
			return
		}

		http.Serve(l, mux)
	}(node2)
	time.Sleep(time.Second)

	err := node1.CallPing(node2.Info)

	if err != nil {
		t.Errorf("ping rpc error: %s", err)
	}

	value := "Hello World"

	// Testing STORE RPC
	err = node1.CallStore(node2.Info, node2.Info.ID, value, false)
	if err != nil {
		t.Errorf("store rpc error: %s", err)
	}

	// Testing FIND_VALUE RPC: should return the value corresponding to the key
	ret, err := node1.CallFindValue(node2.Info, node2.Info.ID)
	if err != nil || ret.Value != value {
		t.Errorf("find value rpc error %s", err)
	}

	// Testing FIND_VALUE RPC: should not return any value back
	ret, err = node1.CallFindValue(node2.Info, node1.Info.ID)
	if err != nil || ret.Value != "" {
		t.Errorf("find value rpc error %s", err)
	}

	// Testing FIND_NODE RPC
	findRet, findErr := node1.CallFindNode(node2.Info)
	if findErr != nil || len(findRet) != 1 || !findRet[0].Equal(node1.Info) {
		t.Errorf("find node rpc error %s", findErr)
	}

}

// Testing recursive lookup operations
func TestNode_Lookups(t *testing.T) {
	serverNodes := 20
	initPort := 5001
	nodes := make([]*Node, 0, serverNodes)

	for i := range serverNodes {
		var id NodeID
		for j := range id {
			id[j] = byte(rand.Intn(256))
		}

		n := MakeNodeStatic(&id, "127.0.0.1", fmt.Sprintf("%d", initPort+i))
		nodes = append(nodes, n)

		go func(n *Node) {
			srv := rpc.NewServer()
			srv.Register(n)

			mux := http.NewServeMux()
			mux.Handle("/_goRPC_", srv)

			port := ":" + n.Info.Port
			l, err := net.Listen("tcp", port)
			if err != nil {
				t.Errorf("tcp listening error: %s", err)
				return
			}

			http.Serve(l, mux)
		}(n)
	}

	time.Sleep(time.Second)

	var ID1 NodeID
	node1 := MakeNodeStatic(&ID1, "127.0.0.1", "5000")

	for _, n := range nodes {
		err := node1.CallPing(n.Info)
		if err != nil {
			t.Errorf("ping rpc error: %s", err)
		}
	}

	value := "Hello World"
	var key NodeID = sha1.Sum([]byte(value))

	pl, _, err := node1.Lookup(&key, 2)

	if err != nil {
		t.Errorf("lookup recursive operation error %s", err)
	}

	for _, item := range pl {
		err := node1.CallStore(item.Info, &key, value, false)
		if err != nil {
			t.Errorf("store rpc after recursive lookup error %s", err)
		}
	}

	retValue, _, err := node1.ValueLookup(&key, 2)
	if err != nil || retValue != value {
		t.Errorf("value lookup recursive operation error %s", err)
	}

}
