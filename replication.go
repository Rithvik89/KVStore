package main

import (
	"bytes"
	"encoding/json"
	"kvstore/utils"
	"log"
	"net/http"
	"sync"
)

func (app *App) ReplicateRecord(rw http.ResponseWriter, r *http.Request) {
	var body WAL
	// Extract the body from the request
	// and unmarshal it into the WriteRecordBody struct
	err := utils.ExtractBody(r, &body)
	if err != nil {
		http.Error(rw, "Failed to extract body", http.StatusBadRequest)
		return
	}

	// append entry to WAL

	app.writeToWAL(body)
	// Store the value in the KV store
	err = app.KvStore.Put(body.Key, body.Value)
	if err != nil {
		http.Error(rw, "Failed to put value", http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (app *App) replicate(opType string, key string, value string, version int32) bool {
	// Create a new WriteRecordBody struct
	body := WAL{
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

	workers, _, err := app.ZkClient.Children("/workers")
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
			workerData, _, err := app.ZkClient.Get("/workers" + worker)
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

	if successCount < app.WriteQuorum {
		log.Println("Failed to replicate to enough workers")
		return false
	}

	return true
}
