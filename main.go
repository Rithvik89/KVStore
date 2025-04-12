package main

import (
	"flag"
	"fmt"
	"kvstore/pkg/kvstore"
	"net/http"
	"sync"
	"time"

	"github.com/apex/log"
	"github.com/go-chi/chi"
	"github.com/go-zookeeper/zk"
)

type App struct {
	KvPort            int              `json:"kv_port"`
	KvStore           kvstore.IKVStore `json:"kv_store"`
	R                 *chi.Mux         `json:"r"`
	ZkClient          *zk.Conn         `json:"zk_client"`
	IsLeader          bool             `json:"is_leader"`
	WriteVersion      int32            `json:"write_version"`
	WriteVersionMutex sync.Mutex       `json:"write_version_mutex"`
	WriteQuorum       int32            `json:"write_quorum"`
	ReadQuorum        int32            `json:"read_quorum"`
	ClusterSize       int32            `json:"cluster_size"`
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
		KvPort:       *port,
		R:            chi.NewRouter(),
		ZkClient:     conn,
		WriteVersion: 1,
	}

	// Leader elections
	app.election()

	// Initialize Cluster Metadata
	app.initializeClusterMetadata()

	// Initialize the KV store
	app.KvStore = kvstore.NewInMemStore()
	// Initialize the WAL LatestSuccessfulWriteVersion
	app.WriteVersion = app.readLatestSuccessfulWriteVersion()

	// Initialize the handler
	app.R.Get("/", app.ReadRecords)
	app.R.Post("/", app.WriteRecord)
	app.R.Delete("/", app.DeleteRecord)

	err = http.ListenAndServe(fmt.Sprintf(":%d", app.KvPort), app.R)

	if err != nil {
		log.Errorf("Failed to start server: %v", err)
	}
}
