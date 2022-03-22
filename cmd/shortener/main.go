package main

import (
	"flag"
	"github.com/caarlos0/env/v6"
	"github.com/go-chi/chi"
	"log"
	"net/http"
)

type Database interface {
	Create(key, value, userID string) error
	Select(key string) (string, error)
	SelectAll(string) map[string]string
}

type Config struct {
	ServerAddress   string `env:"SERVER_ADDRESS" envDefault:"localhost:8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	FileStoragePath string `env:"FILE_STORAGE_PATH"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

func NewRouter(cfg Config) chi.Router {
	r := chi.NewRouter()
	var db Database
	var err error
	if cfg.DatabaseDSN != "" {
		db, err = NewPostgresqlDatabase(cfg)
		if err != nil {
			log.Fatal(err)
		}
	} else if cfg.FileStoragePath != "" {
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
	r.Get("/ping", Ping(cfg))

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
	flag.StringVar(&cfg.DatabaseDSN, "d", cfg.DatabaseDSN, "database dsn")

	flag.Parse()

	r := NewRouter(cfg)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
