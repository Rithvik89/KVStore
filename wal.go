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

func (app *App) writeToWAL(wal WAL) {
	// Open the WAL file for appending
	file, err := os.OpenFile(fmt.Sprintf("wal_%d.log", app.KvPort), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Failed to open WAL file:", err)
		return
	}
	defer file.Close()

	// Increment the write version
	app.WriteVersionMutex.Lock()
	wal.Version = app.WriteVersion
	app.WriteVersion++
	app.WriteVersionMutex.Unlock()

	// Set the success marker
	wal.SuccessMarker = false

	// Write the WAL entry to the file
	err = json.NewEncoder(file).Encode(wal)
	if err != nil {
		log.Println("Failed to write to WAL file:", err)
		return
	}

}

func (app *App) readLatestSuccessfulWriteVersion() int32 {
	// Open the WAL file for reading
	file, err := os.Open(fmt.Sprintf("wal_%d.log", app.KvPort))
	if err != nil {
		log.Println("Failed to open WAL file:", err)
		return 0
	}
	defer file.Close()

	var latestVersion int32 = 1
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

	log.Println("Latest successful write version:", latestVersion)

	return latestVersion
}
