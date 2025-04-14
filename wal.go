package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type WAL struct {
	Version       int32  `json:"version"`
	Type          string `json:"type"`
	Key           string `json:"key"`
	Value         string `json:"value"`
	SuccessMarker bool   `json:"success_marker"`
}

func (app *App) writeToWAL(wal WAL) (int32, error) {
	// Open the WAL file for appending
	file, err := os.OpenFile(fmt.Sprintf("wal_%d.log", app.KvPort), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Failed to open WAL file:", err)
		return -1, err
	}
	defer file.Close()

	// Check for conflicts
	// Get the last successful write version from Zookeeper

	latestVersion := app.readLatestSuccessfulWriteVersionFromWAL()
	latestSuccessfulVersion, err := app.readLastestSuccessfulWriteVersionFromZK()
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

func (app *App) readLatestSuccessfulWriteVersionFromWAL() int32 {
	// Open the WAL file for reading
	file, err := os.Open(fmt.Sprintf("wal_%d.log", app.KvPort))
	if err != nil {
		log.Println("Failed to open WAL file:", err)
		return 0
	}
	defer file.Close()

	var latestVersion int32
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

func (app *App) readLastestSuccessfulWriteVersionFromZK() (int32, error) {
	path := "/version"
	var latestVersion int32
	// Get the latest successful write version from Zookeeper
	data, _, err := app.ZkClient.Get(path)
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
