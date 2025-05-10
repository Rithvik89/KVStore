package main

import (
	"kvstore/internal/wal"
	"kvstore/utils"
	"log"
	"net/http"
)

func (app *App) WALWriter(rw http.ResponseWriter, r *http.Request) {
	// Check if the request is a POST request
	log.Println("Received WAL replication request")
	var body wal.WAL
	// Extract the body from the request
	// and unmarshal it into the WriteRecordBody struct
	err := utils.ExtractBody(r, &body)
	if err != nil {
		http.Error(rw, "Failed to extract body", http.StatusBadRequest)
		return
	}
	if app.ElectionManager.IsLeader {
		http.Error(rw, "Leader cannot replicate", http.StatusBadRequest)
		return
	}

	log.Println("Replicating WAL entry")

	_, err = app.WALManager.WALWriter(wal.WAL{
		Type:          body.Type,
		Key:           body.Key,
		Value:         body.Value,
		SuccessMarker: false,
	})

	if err != nil {
		http.Error(rw, "Failed to replicate WAL to workers", http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (app *App) CommitTxn(rw http.ResponseWriter, r *http.Request) {
	var body wal.WAL
	// Extract the body from the request
	// and unmarshal it into the WriteRecordBody struct
	err := utils.ExtractBody(r, &body)
	if err != nil {
		http.Error(rw, "Failed to extract body", http.StatusBadRequest)
		return
	}

	//TODO: Mark the WAL entry as successful

	err = app.StoreManager.Store.Put(body.Key, body.Value)
	if err != nil {
		http.Error(rw, "Failed to put value", http.StatusInternalServerError)
		return
	}
	rw.WriteHeader(http.StatusOK)

}
