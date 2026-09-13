package main

import (
	"crypto/sha1"
	"fmt"
	"os"
	"time"

	"github.com/Aggelos561/kademlia-dht-go/kademlia"
	"github.com/Aggelos561/kademlia-dht-go/logging"
)

func writeOutput(file *os.File, str string) {
	if file != nil {
		file.WriteString(str)
	}
}

func PingTest(node *kademlia.Node) {
	kclosest := node.FindClosest(node.Info.ID)

	for item := kclosest.Front(); item != nil; item = item.Next() {
		info := item.Value.(*kademlia.NodeInfo)
		err := node.CallPing(info)
		if err != nil {
			logging.Warning("%s", err)
		}
	}
}

func StoreTest(node *kademlia.Node, key, value string, a int, file *os.File) error {
	var hashedKey kademlia.NodeID = sha1.Sum([]byte(key))

	start := time.Now()
	pq, iter, err := node.Lookup(&hashedKey, uint(a))
	elapsed := time.Since(start)

	if err != nil {
		writeOutput(file, fmt.Sprintf("Lookup for key %s failed\n", key))
		return fmt.Errorf("store failed")
	}
	writeOutput(file, fmt.Sprintf("Lookup for key %s succeeded: %s, %d\n", key, elapsed, iter))

	for _, item := range pq {
		err := node.CallStore(item.Info, &hashedKey, value, false)
		if err != nil {
			logging.Warning("%s", err)
			writeOutput(file, fmt.Sprintf("Store for key %s failed\n", key))
		}
	}
	return nil
}

func FindTest(node *kademlia.Node, key string, a int, file *os.File) (string, error) {
	var hashedKey kademlia.NodeID = sha1.Sum([]byte(key))

	start := time.Now()
	value, iter, err := node.ValueLookup(&hashedKey, uint(a))
	elapsed := time.Since(start)

	if err != nil {
		logging.Warning("%s", err)
		writeOutput(file, fmt.Sprintf("ValueLookup for key %s failed\n", key))
		return "", fmt.Errorf("value could not be found")
	} else {
		logging.Info("found value: \"%s\"\n", value)
		writeOutput(file, fmt.Sprintf("ValueLookup for key %s succeeded: %s, %d\n", key, elapsed, iter))
		return value, nil
	}
}
