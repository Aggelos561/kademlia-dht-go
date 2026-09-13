package main

import (
	"bufio"
	"crypto/sha1"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/Aggelos561/kademlia-dht-go/kademlia"
	"github.com/Aggelos561/kademlia-dht-go/logging"
)

var seededRand = rand.New(rand.NewSource(42))
var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func parseLine(line string) (string, string) {
	re := regexp.MustCompile(`^([0-9]{1,3}(?:\.[0-9]{1,3}){3}):(\d+)$`)
	matches := re.FindStringSubmatch(line)
	if matches == nil {
		return "", ""
	}
	ip := matches[1]
	port := matches[2]
	return ip, port
}

func bootstrapping(node *kademlia.Node, bootstrapFile string, delay, a int) {
	time.Sleep(time.Duration(delay) * time.Second)

	file, err := os.Open(bootstrapFile)
	if err != nil {
		return
	}
	defer file.Close()

	bootstrapNodes := make([]*kademlia.NodeInfo, 0)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()

		var ip, port string
		var ID kademlia.NodeID

		ip, port = parseLine(line)

		if node.Info.IP == ip && node.Info.Port == port {
			continue
		}

		ID = sha1.Sum([]byte(ip + ":" + port))
		nodeInfo := &kademlia.NodeInfo{
			ID:   &ID,
			IP:   ip,
			Port: port,
		}
		bootstrapNodes = append(bootstrapNodes, nodeInfo)
	}

	node.Bootstrap(bootstrapNodes, a)
}

func randomStr(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[seededRand.Intn(len(letterRunes))]
	}
	return string(b)
}

func executeExperiment(node *kademlia.Node, keys int, bootstrapFile string, delay, a int, output string) {
	file, err := os.Create(output)
	if err != nil {
		logging.Error("Error creating file: %s", err)
		return
	}
	defer file.Close()

	genKeys := make([]string, 0)
	for range keys {
		node = kademlia.MakeNodeStatic(node.Info.ID, node.Info.IP, node.Info.Port)
		bootstrapping(node, bootstrapFile, delay, a)

		key := randomStr(8)
		value := randomStr(8)
		genKeys = append(genKeys, key)

		StoreTest(node, key, value, a, file)
	}

	for i := range keys {
		node = kademlia.MakeNodeStatic(node.Info.ID, node.Info.IP, node.Info.Port)
		bootstrapping(node, bootstrapFile, delay, a)

		FindTest(node, genKeys[i], a, file)
	}
}

func main() {
	ip := flag.String("ip", "127.0.0.1", "IP to bind")
	port := flag.String("port", "9000", "Port to listen on")
	bootstrapFile := flag.String("file", "", "Bootstrap file with IPs and ports")
	mode := flag.String("mode", "node", "Mode: 'node' or 'client'")
	parallelism := flag.String("a", "1", "Kademlia parallelism parameter 'a'")
	experiment := flag.Bool("experiment", false, "Do an experiment")
	keys := flag.Int("keys", 100, "How many keys to generate during an experiment")
	output := flag.String("output", "logs.txt", "Output file for an experiment")
	delay := flag.Int("delay", 3, "Delay for bootstrapping (seconds)")
	flag.Parse()

	a, err := strconv.Atoi(*parallelism)
	if err != nil || a <= 0 {
		log.Fatal("'a' input parameter must be a decimal number > 0")
	}

	fmt.Printf("Kademlia %s starting [%s:%s]\n", *mode, *ip, *port)
	fmt.Printf("Bootstrap file: %s\n", *bootstrapFile)

	var ID kademlia.NodeID = sha1.Sum([]byte(*ip + ":" + *port))
	node := kademlia.MakeNodeStatic(&ID, *ip, *port)

	if *mode == "client" {
		if *experiment {
			executeExperiment(node, *keys, *bootstrapFile, *delay, a, *output)
		} else {
			bootstrapping(node, *bootstrapFile, *delay, a)
			run_cli(node, a)
		}
	} else {
		go bootstrapping(node, *bootstrapFile, *delay, a)
		go node.KeyRepublishing(1, a)
		go node.RefreshBuckets(1, a)
		node.Listening()
	}
}
