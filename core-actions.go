package main

import (
	"kvstore/utils"
	"net/http"
)

type ReadRecordsBody struct {
	Keys []string `json:"keys"`
}

func (app *App) ReadRecords(rw http.ResponseWriter, r *http.Request) {
	// Implement the logic to get records from the KV store
	var body ReadRecordsBody
	// Extract the body from the request
	// and unmarshal it into the ReadRecordsBody struct
	err := utils.ExtractBody(r, &body)
	if err != nil {
		http.Error(rw, "Failed to extract body", http.StatusBadRequest)
		return
	}
	if !app.IsLeader {
		// Retreive values from the KV store
		outValues := make([]string, len(body.Keys))
		for i, key := range body.Keys {
			value, err := app.KvStore.Get(key)
			if err != nil {
				http.Error(rw, "Failed to get value", http.StatusInternalServerError)
				return
			}
			outValues[i] = value
		}

		// Send the values back to the client
		if err := utils.WriteJSON(rw, outValues); err != nil {
			http.Error(rw, "Failed to write JSON response", http.StatusInternalServerError)
			return
		}
		return
	}

	http.Error(rw, "UnAuthorized action(GET) for a leader ... ", http.StatusForbidden)
}

type WriteRecordBody struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (app *App) WriteRecord(rw http.ResponseWriter, r *http.Request) {
	var body WriteRecordBody
	// Extract the body from the request
	// and unmarshal it into the WriteRecordBody struct
	err := utils.ExtractBody(r, &body)
	if err != nil {
		http.Error(rw, "Failed to extract body", http.StatusBadRequest)
		return
	}
	if app.IsLeader {
		// Store the value in the KV store
		app.writeToWAL(WAL{
			Type:  "PUT",
			Key:   body.Key,
			Value: body.Value,
		})

		// Replicate WAL to followers
		// If acknowledged by more than half of the followers
		// Commit the WAL entry and set the value in the KV store
		// Else rollback the WAL entry
		app.replicate("PUT", body.Key, body.Value)

		err := app.KvStore.Put(body.Key, body.Value)
		if err != nil {
			http.Error(rw, "Failed to put value", http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusOK)
		return
	}

	http.Error(rw, "UnAuthorized action(POST) for a follower ... ", http.StatusForbidden)
}

type DeleteRecordBody struct {
	Key string `json:"key"`
}

func (app *App) DeleteRecord(rw http.ResponseWriter, r *http.Request) {
	var body DeleteRecordBody
	// Extract the body from the request
	// and unmarshal it into the DeleteRecordBody struct
	err := utils.ExtractBody(r, &body)
	if err != nil {
		http.Error(rw, "Failed to extr body", http.StatusBadRequest)
		return
	}
	if app.IsLeader {
		// Delete the value from the KV store
		app.KvStore.Delete(body.Key)
		rw.WriteHeader(http.StatusOK)
		return
	}

	http.Error(rw, "UnAuthorized action(DELETE) for a follower ... ", http.StatusForbidden)
}

func (app *App) ReplicateWAL(rw http.ResponseWriter, r *http.Request) {
	var body WAL
	// Extract the body from the request
	// and unmarshal it into the WriteRecordBody struct
	err := utils.ExtractBody(r, &body)
	if err != nil {
		http.Error(rw, "Failed to extract body", http.StatusBadRequest)
		return
	}
	app.writeToWAL(body)
	app.KvStore.Put(body.Key, body.Value)
	rw.WriteHeader(http.StatusOK)
	// Send a response back to the client
	return
}
