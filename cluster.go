package main

import (
	"fmt"

	"github.com/go-zookeeper/zk"
)

func getClusterSize(app *App) int32 {
	// Get the cluster size from Zookeeper
	path := "/workers" // Zookeeper path for leader election
	// Get the list of children nodes
	children, _, err := app.ZkClient.Get(path)
	if err != nil {
		panic(err)
	}

	return int32(len(children))
}

func getWriteQuorum(app *App) int32 {
	// Calculate the write quorum
	return (app.ClusterSize / 2) + 1
}

func getReadQuorum(app *App) int32 {
	// Calculate the read quorum
	return (app.ClusterSize / 2) + 1
}

func (app *App) initializeClusterMetadata() {
	path := "/workers" // Zookeeper path for leader election

	updateClusterDetails := func() {
		app.ClusterSize = getClusterSize(app)
		app.WriteQuorum = getWriteQuorum(app)
		app.ReadQuorum = getReadQuorum(app)
	}

	updateClusterDetails()

	// Get the list of children nodes
	_, _, ch, err := app.ZkClient.GetW(path)
	if err != nil {
		panic(err)
	}

	for ev := range ch {
		if ev.Type == zk.EventNodeChildrenChanged {
			fmt.Println("Cluster size changed, resetting cluster details")
			updateClusterDetails()
		}
	}
}
