package main

import (
	"github.com/go-chi/chi"
	"io"
	"log"
	"net/http"
	"net/url"
)

func NewRouter() chi.Router {
	r := chi.NewRouter()
	dbm := DatabaseManager{}
	db := MapDatabase{db: &dbm}
	var hashURL Hasing = &Md5HashData{}

	r.Route("/", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			inputURL, err := io.ReadAll(r.Body)
			defer r.Body.Close()

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if _, err := url.ParseRequestURI(string(inputURL)); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			key := hashURL.Hash(inputURL)
			db.Create(key, string(inputURL))

			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(key))
		})

		r.Get("/{ID}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "ID")
			val, err := db.Select(id)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Not found"))
				return
			}
			http.Redirect(w, r, val, http.StatusTemporaryRedirect)
		})
	})

	return r
}

func main() {
	r := NewRouter()
	log.Fatal(http.ListenAndServe(":8080", r))
}
