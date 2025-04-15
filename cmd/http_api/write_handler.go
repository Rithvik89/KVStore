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
		// TODO: this logic need to be moved out from here to wal manager
		// handler should be clean
		app.WALManager.WriteVersionMutex.Lock()
		version := app.WALManager.WriteVersion
		app.WALManager.WriteVersion++
		app.WALManager.WriteVersionMutex.Unlock()

		_, err := app.WALManager.WriteToWAL(wal.WAL{
			Version:       version,
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
		// If acknowledged by more than half of the followers
		// Commit the WAL entry and set the value in the KV store
		// Else rollback the WAL entry

		if !app.ReplicationManager.WriteToWorkers("PUT", body.Key, body.Value, app.WALManager.WriteVersion) {
			// unable to get ack for this WAL entry
			// rollback the WAL entry
			//TODO: visulaize what happens in not all the replicas ack the write
			// app.rollbackWAL(body.Key, body.Value)
		}

		//TODO: Send the Commit to the replicas
		// err = app.CommitToWorkers("PUT",body.Key, body.Value, version)

		// err := app.KvStore.Put(body.Key, body.Value)
		// if err != nil {
		// 	http.Error(rw, "Failed to put value", http.StatusInternalServerError)
		// 	return
		// }

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
