package main

import (
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/go-zookeeper/zk"
)

func (app *App) getChildren() []string {
	path := "/election" // Zookeeper path for leader election
	// Get the list of children nodes
	children, _, err := app.ZkClient.Children(path)
	if err != nil {
		panic(err)
	}
	return children
}

func extractSuffix(node string) string {
	parts := strings.Split(node, "_")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

func (app *App) election() {

	path := "/election/node_"

	// Create a new ephemeral node for this instance
	createPath, err := app.ZkClient.CreateProtectedEphemeralSequential(path, []byte(""), zk.WorldACL(zk.PermAll))
	log.Println("Created node:", createPath)
	if err != nil {
		panic(err)
	}

	for {
		children := app.getChildren()

		sort.Slice(children, func(i, j int) bool {
			return extractSuffix(children[i]) < extractSuffix(children[j])
		})

		log.Println("Children nodes:", children)

		// Check if this instance is the leader by comparing its node with the smallest node
		if createPath == "/election/"+children[0] {
			// This instance is the leader
			app.IsLeader = true
			fmt.Println("This instance is the leader")
			return
		} else {
			// This instance is not the leader
			app.IsLeader = false
			fmt.Println("This instance is not the leader")

			// Register to zookeeper
			workerPath := "/workers/" + extractSuffix(createPath)
			_, err := app.ZkClient.Create(workerPath, []byte(""), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
			if err != nil {
				panic(err)
			}
			fmt.Println("Registered worker:", workerPath)

			// Watch for previous children nodes to see if they are deleted

			index := 0
			for i, child := range children {
				if createPath == "/election/"+child {
					index = i
					break
				}
			}
			predecessor := children[index-1]
			_, _, ch, err := app.ZkClient.GetW("/election/" + predecessor)
			if err != nil {
				panic(err)
			}

			// Watch for changes on the previous node

			select {
			case ev := <-ch:
				if ev.Type == zk.EventNodeDeleted {
					// Predecessor is gone, recheck election
					continue
				}
			case <-time.After(10 * time.Second):
				// Optional timeout to avoid getting stuck
				fmt.Println("Timeout while waiting, rechecking election.")
				continue
			}
			// Watch the previous node

		}
	}

}
