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

func NewRouter(cfg config.Config, db databases.Database) chi.Router {
	r := chi.NewRouter()
	hashURL := &datahashes.Md5HashData{}

	r.Use(middleware.Logger)
	r.Use(middlewares.CookieMiddleware)
	r.Use(middlewares.GzipRequestMiddleware)
	r.Use(middlewares.GzipResponseMiddleware)

	r.Post("/", handlers.GenerateShortURL(hashURL, db, cfg))
	r.Get("/{ID}", handlers.RedirectFromShortToFull(db))
	r.Post("/api/shorten", handlers.GenerateShortenJSONURL(hashURL, db, cfg))
	r.Get("/api/user/urls", handlers.GetAllShortenURLS(db, cfg))
	r.Get("/ping", handlers.Ping(db))
	r.Post("/api/shorten/batch", handlers.GenerateManyShortenJSONURL(hashURL, db, cfg))
	r.Delete("/api/user/urls", handlers.Delete(db))

	return r
}

func main() {
	var cfg config.Config
	if err := cfg.Parse(); err != nil {
		log.Fatal(err)
	}

	var db databases.Database
	var err error
	if cfg.DatabaseDSN != "" {
		db, err = databases.NewPostgresqlDatabase(cfg)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
	} else if cfg.FileStoragePath != "" {
		db, err = databases.NewFileDatabase(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
	} else {
		db = databases.NewMapDatabase()
		defer db.Close()
	}

	r := NewRouter(cfg, db)
	log.Fatal(http.ListenAndServe(cfg.ServerAddress, r))
}
