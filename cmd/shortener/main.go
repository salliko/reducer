package main

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/salliko/reducer/config"
	"github.com/salliko/reducer/internal/databases"
	"github.com/salliko/reducer/internal/datahashes"
	"github.com/salliko/reducer/internal/handlers"
	"github.com/salliko/reducer/internal/middlewares"
	"log"
	"net/http"
)

func NewRouter(cfg config.Config) chi.Router {
	r := chi.NewRouter()
	var db databases.Database
	var err error
	if cfg.DatabaseDSN != "" {
		db, err = databases.NewPostgresqlDatabase(cfg)
		if err != nil {
			log.Fatal(err)
		}
	} else if cfg.FileStoragePath != "" {
		db, err = databases.NewFileDatabase(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		db = databases.NewMapDatabase()
	}
	hashURL := &datahashes.Md5HashData{}

	r.Use(middleware.Logger)
	r.Use(middlewares.CookieMiddleware)
	r.Use(middlewares.GzipMiddleware)

	r.Post("/", handlers.GenerateShortURL(hashURL, db, cfg))
	r.Get("/{ID}", handlers.RedirectFromShortToFull(db))
	r.Post("/api/shorten", handlers.GenerateShortenJSONURL(hashURL, db, cfg))
	r.Get("/api/user/urls", handlers.GetAllShortenURLS(db, cfg))
	r.Get("/ping", handlers.Ping(cfg))
	r.Post("/api/shorten/batch", handlers.GenerateManyShortenJSONURL(hashURL, db, cfg))

	return r
}

func main() {
	var cfg config.Config
	if err := cfg.Parse(); err != nil {
		log.Fatal(err)
	}

	r := NewRouter(cfg)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
