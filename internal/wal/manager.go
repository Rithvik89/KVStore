package wal

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/go-zookeeper/zk"
)

type WALManager struct {
	KvPort            int        `json:"kv_port"`
	ZkClient          *zk.Conn   `json:"zk_client"`
	WriteVersion      int        `json:"write_version"`
	WriteVersionMutex sync.Mutex `json:"write_version_mutex"`
}

func NewWALManager(kv_port int, zkClient *zk.Conn) *WALManager {
	return &WALManager{
		KvPort:       kv_port,
		ZkClient:     zkClient,
		WriteVersion: readLatestSuccessfulWriteVersionFromWAL(kv_port),
	}
}

type WAL struct {
	Version       int    `json:"version"`
	Type          string `json:"type"`
	Key           string `json:"key"`
	Value         string `json:"value"`
	SuccessMarker bool   `json:"success_marker"`
}

func (wm *WALManager) WALWriter(wal WAL) (int, error) {
	// Open the WAL file for appending
	file, err := os.OpenFile(fmt.Sprintf("wal_%d.log", wm.KvPort), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Failed to open WAL file:", err)
		return -1, err
	}
	defer file.Close()

	// Increment the write version
	wm.WriteVersionMutex.Lock()
	wal.Version = wm.WriteVersion
	wm.WriteVersion++
	wm.WriteVersionMutex.Unlock()

	// Check for conflicts
	conflictDetected, err := wm.isConflictDetected()
	if err != nil {
		log.Println("Error while checking for conflicts:", err)
		return -1, err
	}
	if conflictDetected {
		log.Println("Conflict detected: WAL entry not written")
		//TODO: add conflict resolution logic here
		return -1, fmt.Errorf("conflict detected: WAL entry not written")
	}

	// Write the WAL entry to the file
	err = json.NewEncoder(file).Encode(wal)
	if err != nil {
		log.Println("Failed to write to WAL file:", err)
		return -1, err
	}
	return wal.Version, nil
}

func (wm *WALManager) isConflictDetected() (bool, error) {
	latestVersion := readLatestSuccessfulWriteVersionFromWAL(wm.KvPort)
	latestSuccessfulVersion, err := readLastestSuccessfulWriteVersionFromZK(wm.ZkClient)
	if err != nil {
		log.Println("Failed to get latest successful write version from Zookeeper:", err)
		return false, err
	}
	if latestSuccessfulVersion != latestVersion {
		// Conflict detected
		log.Println("Conflict detected: latest successful version from Zookeeper does not match WAL version")
		return true, nil
	}
	return false, nil
}

func readLatestSuccessfulWriteVersionFromWAL(KvPort int) int {
	// Open the WAL file for reading
	file, err := os.Open(fmt.Sprintf("wal_%d.log", KvPort))
	if err != nil {
		log.Println("Failed to open WAL file:", err)
		return 0
	}
	defer file.Close()

	var latestVersion int
	var wal WAL

	// Read the WAL entries from the file
	for {
		err = json.NewDecoder(file).Decode(&wal)
		if err != nil {
			break
		}
		//TODO: This should only happen if there is a success marker
		latestVersion = wal.Version
	}
	return latestVersion
}

func readLastestSuccessfulWriteVersionFromZK(ZkClient *zk.Conn) (int, error) {
	path := "/version"
	var latestVersion int
	// Get the latest successful write version from Zookeeper
	data, _, err := ZkClient.Get(path)
	if err != nil {
		log.Println("Failed to get latest successful write version from Zookeeper:", err)
		return -1, err
	}
	err = json.Unmarshal(data, &latestVersion)
	if err != nil {
		log.Println("Failed to unmarshal latest successful write version:", err)
		return -1, err
	}
	return latestVersion, nil
}
