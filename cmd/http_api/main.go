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
		Handler:        chi.NewRouter(),
		ClusterManager: cluster.NewClusterManager(*port, conn),
		StoreManager:   store.NewStoreManager(),
	}

	app.ElectionManager = elections.NewElectionManager(*port, conn)
	fmt.Println("Election Manager initialized")
	app.ElectionManager.Election()

	// Intialize WAL manager
	app.WALManager = wal.NewWALManager(*port, conn)
	fmt.Println("WAL Manager initialized")

	// Initialize Replication Manager
	app.ReplicationManager = replication.NewReplicationManager(*port, conn, app.WALManager, app.ClusterManager)
	fmt.Println("Replication Manager initialized")

	// Initialize Cluster Metadata
	app.ClusterManager.InitializeClusterMetadata()

	// Initialize Handler
	app.InitializeHandler()

	fmt.Println("KV Store is running on port:", *port)
	
	err = http.ListenAndServe(fmt.Sprintf(":%d", *port), app.Handler)

	if err != nil {
		log.Errorf("Failed to start server: %v", err)
	}
}
