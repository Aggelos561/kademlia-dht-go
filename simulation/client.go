package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/Aggelos561/kademlia-dht-go/kademlia"
	"github.com/Aggelos561/kademlia-dht-go/logging"
)

func run_cli(node *kademlia.Node, a int) {

	logging.EnableStdErr()

	for {
		fmt.Printf(">> ")
		reader := bufio.NewReader(os.Stdin)
		command, _ := reader.ReadString('\n')

		cmdTokens := strings.Fields(command)
		if len(cmdTokens) == 0 {
			continue
		}

		operation := cmdTokens[0]

		if operation == "store" {
			if len(cmdTokens) != 3 {
				fmt.Println("invalid \"store\" format!")
				continue
			}
			key, value := cmdTokens[1], cmdTokens[2]
			err := StoreTest(node, key, value, a, nil)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("store success")
			}
		} else if operation == "find" {
			if len(cmdTokens) != 2 {
				fmt.Println("invalid \"find\" format!")
				continue
			}
			key := cmdTokens[1]
			value, err := FindTest(node, key, a, nil)
			if err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Value: ", value)
			}

		} else if operation == "quit" {
			return
		}
	}
}
