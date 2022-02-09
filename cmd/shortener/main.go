package main

import (
	"fmt"
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
			newURL := fmt.Sprintf("http://localhost:8080/%s", key)
			w.Write([]byte(newURL))
		})

		r.Get("/{ID}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "ID")
			val, err := db.Select(id)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Not found"))
				return
			}
			w.Header().Set("Location", val)
			w.WriteHeader(http.StatusTemporaryRedirect)
			//http.Redirect(w, r, val, http.StatusTemporaryRedirect)
			w.Write([]byte("Found"))
		})
	})

	return r
}

func main() {
	r := NewRouter()
	log.Fatal(http.ListenAndServe(":8080", r))
}
