package replication

import (
	"bytes"
	"encoding/json"
	"fmt"
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

func (rm *ReplicationManager) WALReplicationToWorkers(opType string, key string, value string, version int) error {
	body := wal.WAL{
		Version: version,
		Type:    opType,
		Key:     key,
		Value:   value,
	}

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
			resp, err := http.Post("http://"+workerAddress+"/api/v1/replicate/", "application/json", bytes.NewBuffer(bodyJson))
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
		return fmt.Errorf("failed to replicate to enough workers: %d/%d", successCount, rm.ClusterManager.WriteQuorum)
	}

	return nil
}

func (rm *ReplicationManager) CommitTxnToWorkers(opType string, key string, value string, version int) error {
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
