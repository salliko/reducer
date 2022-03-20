package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi"
	"log"
	"net/http"
)

type Database interface {
	Create(userID, key, value string) error
	Select(userID, key string) (string, error)
	SelectAll(string) map[string]string
}

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
}

func NewRouter(cfg Config) chi.Router {
	r := chi.NewRouter()
	var db Database
	var err error
	if cfg.FileStoragePath != "" {
		db, err = NewFileDatabase(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		db = NewMapDatabase()
	}
	hashURL := &Md5HashData{}

	r.Use(CookieMiddleware)
	r.Use(GzipMiddleware)

	r.Post("/", GenerateShortURL(hashURL, db, cfg))
	r.Get("/{ID}", RedirectFromShortToFull(db))
	r.Post("/api/shorten", GenerateShortenJSONURL(hashURL, db, cfg))
	r.Get("/api/user/urls", GetAllShortenURLS(db, cfg))

	return r
}

func main() {
	var cfg Config
	err := env.Parse(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	flag.StringVar(&cfg.ServerAddress, "a", cfg.ServerAddress, "server address")
	flag.StringVar(&cfg.BaseURL, "b", cfg.BaseURL, "base url")
	flag.StringVar(&cfg.FileStoragePath, "f", cfg.FileStoragePath, "file storage path")

	flag.Parse()

	r := NewRouter(cfg)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
