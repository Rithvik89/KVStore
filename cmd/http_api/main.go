package main

import (
	"flag"
	"fmt"
	"kvstore/internal/cluster"
	"kvstore/internal/elections"
	store "kvstore/internal/kv"
	"kvstore/internal/replication"
	"kvstore/internal/wal"

	"net/http"
	"time"

	"github.com/apex/log"
	"github.com/go-chi/chi"
	"github.com/go-zookeeper/zk"
)

type App struct {
	Handler            *chi.Mux                        `json:"handler"`
	ClusterManager     *cluster.ClusterManager         `json:"cluster_manager"`
	ElectionManager    *elections.ElectionManager      `json:"election_manager"`
	ReplicationManager *replication.ReplicationManager `json:"replication_manager"`
	WALManager         *wal.WALManager                 `json:"wal_manager"`
	StoreManager       *store.StoreManager             `json:"store_manager"`
}

func main() {

	// The port variable is a pointer to an int that will hold the value of the port flag after parsing.
	port := flag.Int("port", 8081, "Port for the KV store")
	// here the value will be loaded into the port variable..
	flag.Parse()

	// Connect to Zookeeper
	conn, _, err := zk.Connect([]string{"localhost:2181"}, 5*time.Second)

	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Initialize the application
	app := App{
		ElectionManager:    elections.NewElectionManager(*port, conn),
		ClusterManager:     cluster.NewClusterManager(*port, conn),
		ReplicationManager: replication.NewReplicationManager(*port, conn),
		WALManager:         wal.NewWALManager(*port, conn),
		StoreManager:       store.NewStoreManager(),
	}

	// Leader elections
	app.ElectionManager.Election()

	// Initialize Cluster Metadata
	app.ClusterManager.InitializeClusterMetadata()

	err = http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)

	if err != nil {
		log.Errorf("Failed to start server: %v", err)
	}
}
