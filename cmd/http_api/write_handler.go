package main

import (
	"kvstore/internal/wal"
	"kvstore/utils"
	"net/http"
)

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
	if app.ElectionManager.IsLeader {
		
		// 2PC Prepare Phase
		version, err := app.WALManager.WALWriter(wal.WAL{
			Type:          "PUT",
			Key:           body.Key,
			Value:         body.Value,
			SuccessMarker: false,
		})
		if err != nil {
			http.Error(rw, "Failed to write to WAL", http.StatusInternalServerError)
			return
		}

		// Replicate WAL to followers
		err = app.ReplicationManager.WALReplicationToWorkers("PUT", body.Key, body.Value, version)
		if err != nil {
			// These false WAL entries will be cleaned up during compaction
			http.Error(rw, "Failed to replicate WAL to workers", http.StatusInternalServerError)
			return
		}

		// 2PC Commit Phase
		
		err = app.ReplicationManager.CommitTxnToWorkers("PUT", body.Key, body.Value, version)
		if err != nil {
			http.Error(rw, "Failed to commit transaction to workers", http.StatusInternalServerError)
			return
		}

		err = app.StoreManager.Store.Put(body.Key, body.Value)
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
	if app.ElectionManager.IsLeader {
		// Delete the value from the KV store
		app.StoreManager.Store.Delete(body.Key)
		rw.WriteHeader(http.StatusOK)
		return
	}

	http.Error(rw, "UnAuthorized action(DELETE) for a follower ... ", http.StatusForbidden)
}
