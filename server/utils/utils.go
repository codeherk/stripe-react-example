package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
)

func WriteJSON(w http.ResponseWriter, v interface{}) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewEncoder.Encode: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.Copy(w, &buf); err != nil {
		log.Printf("io.Copy: %v", err)
		return
	}
}

func GetEnvOrDefault(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		log.Printf(`%s not set. Defaulting to "%s".`, key, defaultValue)
		return defaultValue
	}
	log.Printf(`%s set to "%s".`, key, value)
	return value
}
