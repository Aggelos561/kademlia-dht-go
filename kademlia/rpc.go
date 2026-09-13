package kademlia

import (
	"container/list"
	"database/sql"
	"encoding/hex"
	"math"
	"time"

	"github.com/Aggelos561/kademlia-dht-go/logging"

	_ "modernc.org/sqlite"
)

const NETSIM = true

type StoreArgs struct {
	Sender  *NodeInfo
	Key     *NodeID
	Value   string
	Caching bool
}

type FindNodeArgs struct {
	Sender *NodeInfo
	ID     *NodeID
}

type PingArgs struct {
	Sender *NodeInfo
}

type FindValueArgs struct {
	Sender *NodeInfo
	ID     *NodeID
}

type FindValueRet struct {
	Nodes []*NodeInfo
	Value string
}

// Converts a linked list of NodeInfo pointers to a slice of NodeInfo pointers.
func listToSlice(l *list.List, n *[]*NodeInfo) {
	for e := l.Front(); e != nil; e = e.Next() {
		*n = append(*n, e.Value.(*NodeInfo))
	}
}

// Calculate the expiration time for a key-value pair.
func (node *Node) expirationTime(key NodeID, caching bool) int64 {
	current_time := time.Now().Unix()
	if !caching {
		return current_time + int64(24*time.Hour.Seconds())
	}
	bucket, _ := FindIndex(node.Info.ID, &key)
	normalizedBucket := float64(16 - int(math.Floor(float64(bucket)*15.0/159.0)) + 1)
	return current_time + int64((0.66)*math.Pow(2, float64(normalizedBucket)))
}

// Pong an incoming ping RPC
func (node *Node) Ping(args *PingArgs, reply *any) error {

	sender := args.Sender
	node.PingHandler(sender)
	logging.Info("[receiver] Ping RPC [%s:%s → %s:%s]", sender.IP, sender.Port, node.Info.IP, node.Info.Port)
	return nil
}

// Find the K closest nodes to a given node and return via reply
func (node *Node) FindNode(args *FindNodeArgs, reply *[]*NodeInfo) error {
	sender, id := args.Sender, args.ID
	kclosest := node.FindClosest(id)
	listToSlice(kclosest, reply)

	// Simulate network delay
	if NETSIM {
		time.Sleep(time.Second)
	}

	node.PingHandler(sender)
	logging.Info("[receiver] FindNode RPC [%s:%s → %s:%s] for key %x", sender.IP, sender.Port, node.Info.IP, node.Info.Port, *id)

	return nil
}

// Store a key-value pair in the node's sqlite local database.
func (node *Node) Store(args *StoreArgs, reply *any) error {
	sender, key, value, caching := args.Sender, *args.Key, args.Value, args.Caching

	stmt, err := node.DB.Prepare(`
		INSERT INTO keys (key, value, expiration)
		VALUES (?, ?, ?)
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			expiration = excluded.expiration;
		`)
	if err != nil {
		logging.Error("%s", err)
		stmt.Close()
		return err
	}
	defer stmt.Close()

	node.dbMux.Lock()
	_, err = stmt.Exec(hex.EncodeToString(key[:]), value, node.expirationTime(key, caching))
	node.dbMux.Unlock()

	if err != nil {
		logging.Error("%s", err)
		return err
	}

	node.PingHandler(sender)
	logging.Info("[receiver] Store RPC [%s:%s → %s:%s]: stored value \"%s\" for key %x", sender.IP, sender.Port, node.Info.IP, node.Info.Port, value, key)
	return nil
}

// Retrieve the value associated with a given key, if it exists, and return it alongside the key's K closest nodes.
func (node *Node) FindValue(args *FindValueArgs, reply *FindValueRet) error {
	sender, id := args.Sender, args.ID
	kclosest := node.FindClosest(id)
	listToSlice(kclosest, &reply.Nodes)

	node.dbMux.Lock()
	row := node.DB.QueryRow("SELECT key, value, expiration FROM keys WHERE key = ?", hex.EncodeToString(id[:]))
	node.dbMux.Unlock()

	var keyStr string
	var value string
	var expiration int64

	err := row.Scan(&keyStr, &value, &expiration)

	if err != nil {
		if err == sql.ErrNoRows {
			reply.Value = ""
			logging.Info("[receiver] FindValue RPC [%s:%s → %s:%s]: value not found for key %x", sender.IP, sender.Port, node.Info.IP, node.Info.Port, *id)
		} else {
			logging.Error("%s", err)
		}
	} else if expiration < time.Now().Unix() {
		reply.Value = ""
		logging.Info("[receiver] FindValue RPC [%s:%s → %s:%s]: value expired for key %x", sender.IP, sender.Port, node.Info.IP, node.Info.Port, *id)
	} else {
		reply.Value = value
		logging.Info("[receiver] FindValue RPC [%s:%s → %s:%s]: found value \"%s\" for key %x", sender.IP, sender.Port, node.Info.IP, node.Info.Port, value, *id)
	}

	// Simulate network delay
	if NETSIM {
		time.Sleep(time.Second)
	}

	node.PingHandler(sender)
	return nil
}
