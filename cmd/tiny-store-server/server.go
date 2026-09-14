package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"tiny-store/internal/storage"
)

type server struct {
	store storage.Storage
}

type putRequest struct {
	Value string `json:"value"`
}

type getResponse struct {
	Value string `json:"value"`
}

type listResponse struct {
	Entries map[string]string `json:"entries"`
}

func (s *server) put(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	const maxBytes = 1048576

	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	var putReq putRequest

	if err := json.NewDecoder(r.Body).Decode(&putReq); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			http.Error(w, "Body too large", http.StatusRequestEntityTooLarge)
			return
		}

		http.Error(w, "Error reading body", http.StatusBadRequest)
		return
	}

	if err := s.store.Put(key, putReq.Value); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *server) get(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	value, err := s.store.Get(key)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, "Key not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, getResponse{Value: value})
}

func (s *server) delete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")

	if err := s.store.Delete(key); err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			http.Error(w, "Key not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *server) list(w http.ResponseWriter, r *http.Request) {
	entries := s.store.List()

	writeJSON(w, listResponse{Entries: entries})
}

func writeJSON(w http.ResponseWriter, response any) {
	responseJSON, err := json.Marshal(response)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(responseJSON); err != nil {
		log.Printf("write response %v", err)
	}
}
