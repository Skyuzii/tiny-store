package main

import (
	"net/http"
	"tiny-store/internal/storage"
)

func main() {
	server := &server{store: storage.NewMemoryStorage()}

	mux := http.NewServeMux()
	mux.HandleFunc("PUT /kv/{key}", server.put)
	mux.HandleFunc("GET /kv/{key}", server.get)
	mux.HandleFunc("DELETE /kv/{key}", server.delete)
	mux.HandleFunc("GET /kv", server.list)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
