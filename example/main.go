package main

import (
	"fmt"
	"net/http"

	"github.com/bastengao/go-storage"
)

var store storage.Storage

func main() {
	dirPath := "./example/tmp"
	service, err := storage.NewDiskService(dirPath, "http://localhost:8080/disk")
	if err != nil {
		panic(err)
	}

	store = storage.New(service, nil)

	server := storage.NewServer("http://localhost:8080/storage/redirect", store, nil, nil)

	http.Handle("/disk/", storage.ServeDisk("/disk/", dirPath))
	http.Handle("/storage/redirect", server.Handler())
	http.Handle("/upload", http.HandlerFunc(upload))
	http.Handle("/delete", http.HandlerFunc(delete))

	fmt.Println("Listening on http://127.0.0.1:8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func upload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(32 << 20)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer file.Close()

	err = store.Service().Upload(r.Context(), "test.png", file)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func delete(w http.ResponseWriter, r *http.Request) {
	err := store.Service().Delete(r.Context(), "test.png")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
