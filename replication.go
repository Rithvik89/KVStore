package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

func (app *App) replicate(opType string, key string, value string) bool {
	// Create a new WriteRecordBody struct
	body := WAL{
		Type:  opType,
		Key:   key,
		Value: value,
	}

	// Marshal the body into JSON
	bodyJson, err := json.Marshal(body)
	if err != nil {
		log.Println("Failed to marshal body:", err)
		// TODO: Is this way of hadling error correct?
		return false
	}

	workers, _, err := app.ZkClient.Children("/workers")

	// Send the JSON to all followers

	for _, worker := range workers {
		go func(worker string) {
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

			resp.Body.Close()
		}(worker)
	}

	return true
}
