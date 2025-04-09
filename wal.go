package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type WAL struct {
	Type  string `json:"type"`
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (app *App) writeToWAL(wal WAL) {
	// Open the WAL file for appending
	file, err := os.OpenFile(fmt.Sprintf("wal_%d.log", app.KvPort), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Failed to open WAL file:", err)
		return
	}
	defer file.Close()

	// Write the WAL entry to the file
	err = json.NewEncoder(file).Encode(wal)
	if err != nil {
		log.Println("Failed to write to WAL file:", err)
		return
	}
}
