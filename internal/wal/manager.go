package wal

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/go-zookeeper/zk"
)

type WALManager struct {
	KvPort       int      `json:"kv_port"`
	ZkClient     *zk.Conn `json:"zk_client"`
	WriteVersion int      `json:"write_version"`
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

func (wm *WALManager) WriteToWAL(wal WAL) (int, error) {
	// Open the WAL file for appending
	file, err := os.OpenFile(fmt.Sprintf("wal_%d.log", wm.KvPort), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Failed to open WAL file:", err)
		return -1, err
	}
	defer file.Close()

	// Check for conflicts
	// Get the last successful write version from Zookeeper

	latestVersion := readLatestSuccessfulWriteVersionFromWAL(wm.KvPort)
	latestSuccessfulVersion, err := wm.readLastestSuccessfulWriteVersionFromZK()
	if err != nil {
		log.Println("Failed to get latest successful write version from Zookeeper:", err)
		return -1, err
	}

	if latestSuccessfulVersion != latestVersion {
		// Conflict detected
		log.Println("Conflict detected: latest successful version from Zookeeper does not match WAL version")
		//TODO: Conflict resolution logic can be added here
		return -1, fmt.Errorf("conflict detected: latest successful version from Zookeeper does not match WAL version")

	}

	// Write the WAL entry to the file
	err = json.NewEncoder(file).Encode(wal)
	if err != nil {
		log.Println("Failed to write to WAL file:", err)
		return -1, err
	}
	return wal.Version, nil
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

func (wm *WALManager) readLastestSuccessfulWriteVersionFromZK() (int, error) {
	path := "/version"
	var latestVersion int
	// Get the latest successful write version from Zookeeper
	data, _, err := wm.ZkClient.Get(path)
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
