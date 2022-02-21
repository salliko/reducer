package main

import (
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi"
	"log"
	"net/http"
)

const (
	host         = "http://localhost"
	port         = ":8080"
	fullHostPath = host + port
)

var (
	ServerAddress string
	BaseURL       string
)

type Config struct {
	ServerAddress string `env:"SERVER_ADDRESS"`
	BaseURL       string `env:"BASE_URL"`
}

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
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	if cfg.ServerAddress != "" {
		ServerAddress = cfg.ServerAddress
	} else {
		ServerAddress = fullHostPath
	}

	if cfg.BaseURL != "" {
		BaseURL = cfg.BaseURL
	} else {
		BaseURL = fullHostPath
	}

	r := NewRouter()
	log.Fatal(http.ListenAndServe(ServerAddress, r))
}
