package kademlia

import (
	"database/sql"
	"encoding/hex"
	"log"
	"os"
	"time"

	"github.com/Aggelos561/kademlia-dht-go/logging"
)

// Initialize a new SQLite database for the given NodeID
func MakeDB(id NodeID) *sql.DB {
	err := os.MkdirAll("./dbs", 0755)
	if err != nil {
		logging.Warning("Failed to create database directory: %s", err)
	}

	dbname := "./dbs/db" + hex.EncodeToString(id[:]) + ".sqlite"

	db, err := sql.Open("sqlite", dbname)
	if err != nil {
		log.Fatal(err)
	}

	sqlStmt := `CREATE TABLE IF NOT EXISTS keys (
			key TEXT NOT NULL PRIMARY KEY,
			value TEXT NOT NULL,
			expiration INTEGER NOT NULL
			);`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("Create table failed: %s", err)
	}
	return db
}

// Flush the cache by deleting expired keys from the database
func flushCache(node *Node) {
	node.dbMux.Lock()
	_, err := node.DB.Exec("DELETE FROM keys WHERE expiration < ?", time.Now().Unix())

	if err != nil {
		logging.Warning("%s", err)
	} else {
		logging.Info("cache flushed successfully")
	}
	node.dbMux.Unlock()
}

// Republish keys to the closest nodes based on the given NodeID and replication factor
func republishKeys(node *Node, a int) {
	node.dbMux.Lock()
	rows, err := node.DB.Query("SELECT key, value, expiration FROM keys WHERE expiration > ?", time.Now().Unix())

	if err != nil {
		logging.Warning("%s", err)
	}

	defer node.dbMux.Unlock()
	defer rows.Close()

	for rows.Next() {
		var key, value string
		var expiration int64
		var id NodeID

		err = rows.Scan(&key, &value, &expiration)

		if err != nil {
			logging.Warning("%s", err)
			continue
		}

		decodedKey, _ := hex.DecodeString(key)
		copy(id[:], decodedKey)

		pq, _, err := node.Lookup(&id, uint(a))
		if err != nil {
			logging.Warning("%s", err)
			continue
		}

		for _, item := range pq {
			err := node.CallStore(item.Info, &id, value, false)
			if err != nil {
				logging.Warning("%s", err)
			} else {
				logging.Info("key %s republished to [%s:%s]", key, item.Info.IP, item.Info.Port)
			}
		}
	}
}

// Key republishing daemon that periodically flushes the cache and republishes keys
func (node *Node) KeyRepublishing(hours uint, a int) {
	for {
		time.Sleep(time.Duration(hours) * time.Hour)
		flushCache(node)
		republishKeys(node, a)
	}
}
