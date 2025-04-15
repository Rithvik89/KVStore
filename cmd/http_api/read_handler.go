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
	if !app.ElectionManager.IsLeader {
		// Retreive values from the KV store
		outValues := make([]string, len(body.Keys))
		for i, key := range body.Keys {
			value, err := app.StoreManager.Store.Get(key)
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
