package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi"
	"log"
	"net/http"
)

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func NewRouter(cfg Config) chi.Router {
	r := chi.NewRouter()
	var db Database
	if cfg.FileStoragePath != "" {
		db = NewFileDatabase(cfg.FileStoragePath)
	} else {
		db = NewMapDatabase()
	}
	hashURL := &Md5HashData{}

	r.Post("/", GenerateShortURL(hashURL, db, cfg))
	r.Get("/{ID}", RedirectFromShortToFull(db))
	r.Post("/api/shorten", GenerateShortenJSONURL(hashURL, db, cfg))

	return r
}

func main() {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	r := NewRouter(cfg)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
