package replication

import (
	"bytes"
	"encoding/json"
	"kvstore/internal/cluster"
	"kvstore/internal/wal"
	"log"
	"net/http"
	"sync"

	"github.com/go-zookeeper/zk"
)

type ReplicationManager struct {
	KvPort         int                     `json:"kv_port"`
	ZkClient       *zk.Conn                `json:"zk_client"`
	WALManager     *(wal.WALManager)       `json:"wal_manager"`
	ClusterManager *cluster.ClusterManager `json:"cluster_manager"`
}

func NewReplicationManager(kvPort int, zkClient *zk.Conn) *ReplicationManager {
	return &ReplicationManager{
		ZkClient:       zkClient,
		WALManager:     wal.NewWALManager(int(kvPort), zkClient),
		ClusterManager: cluster.NewClusterManager(int(kvPort), zkClient),
	}
}

// TODO: Rename this function to something more meaningful
func (rm *ReplicationManager) WriteToWorkers(opType string, key string, value string, version int) bool {
	// Create a new WriteRecordBody struct

	body := wal.WAL{
		Version: version,
		Type:    opType,
		Key:     key,
		Value:   value,
	}

	// Marshal the body into JSON
	bodyJson, err := json.Marshal(body)
	if err != nil {
		log.Println("Failed to marshal body:", err)
		// TODO: Is this way of hadling error correct?
		return false
	}

	workers, _, err := rm.ZkClient.Children("/workers")
	if err != nil {
		log.Println("Failed to get workers:", err)
		return false
	}

	// Send the JSON to all followers

	wg := sync.WaitGroup{}
	successCount := int32(0)
	mu := sync.Mutex{}
	wg.Add(len(workers))

	for _, worker := range workers {
		go func(worker string) {
			defer wg.Done()
			workerData, _, err := rm.ZkClient.Get("/workers" + worker)
			if err != nil {
				log.Println("Failed to get worker data:", err)
				return
			}
			workerAddress := string(workerData)
			resp, err := http.Post("http://"+workerAddress+"/replica/", "application/json", bytes.NewBuffer(bodyJson))
			if err != nil {
				log.Println("Failed to send replication request:", err)
				return
			}
			// validate the ack from the workers ...
			if resp.StatusCode == http.StatusOK {
				mu.Lock()
				successCount++
				mu.Unlock()
			}
			resp.Body.Close()
		}(worker)
	}

	wg.Wait()

	if successCount < rm.ClusterManager.WriteQuorum {
		log.Println("Failed to replicate to enough workers")
		return false
	}

	return true
}

func (rm *ReplicationManager) CommitToWorkers(opType string, key string, value string, version int) error {
	// Create a new WriteRecordBody struct
	body := wal.WAL{
		Type:    opType,
		Version: version,
		Key:     key,
		Value:   value,
	}

	// Marshal the body into JSON
	bodyJson, err := json.Marshal(body)
	if err != nil {
		log.Println("Failed to marshal body:", err)
		return err
	}

	workers, _, err := rm.ZkClient.Children("/workers")
	if err != nil {
		log.Println("Failed to get workers:", err)
		return err
	}

	// Send the Commit on the version to all followers

	for _, worker := range workers {
		workerData, _, err := rm.ZkClient.Get("/workers" + worker)
		if err != nil {
			log.Println("Failed to get worker data:", err)
			return err
		}
		workerAddress := string(workerData)
		resp, err := http.Post("http://"+workerAddress+"/commit/", "application/json", bytes.NewBuffer(bodyJson))
		if err != nil {
			log.Println("Failed to send replication request:", err)
			return err
		}
		resp.Body.Close()
	}

	return nil

}
