package utils

import (
	"encoding/json"
	"io"
	"net/http"
)

func ExtractBody(r *http.Request, v interface{}) error {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &v)
	if err != nil {
		return err
	}

	return nil
}

func WriteJSON(rw http.ResponseWriter, v interface{}) error {
	rw.Header().Set("Content-Type", "application/json")
	rw.WriteHeader(http.StatusOK)
	body, err := json.Marshal(v)
	if err != nil {
		return err
	}
	_, err = rw.Write(body)
	if err != nil {
		return err
	}
	return nil
}
