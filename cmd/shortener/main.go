package main

import (
	"github.com/go-chi/chi"
	"log"
	"net/http"
)

const (
	host         = "http://localhost"
	port         = ":8080"
	fullHostPath = host + port
)

func NewRouter() chi.Router {
	r := chi.NewRouter()
	db := NewMapDatabase()
	hashURL := &Md5HashData{}

	r.Post("/", GenerateShortURL(hashURL, db))
	r.Get("/{ID}", RedirectFromShortToFull(db))
	r.Post("/api/shorten", GenerateShortenJSONURL(hashURL, db))

	return r
}

func main() {
	r := NewRouter()
	log.Fatal(http.ListenAndServe(port, r))
}
