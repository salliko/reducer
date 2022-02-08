package main

import (
	"crypto/md5"
	"fmt"
	"github.com/go-chi/chi"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

func Normalize(hash string) string {
	return strings.ReplaceAll(hash, "/", "X")
}

func HashURL(url []byte, short bool) string {
	hash := fmt.Sprintf("%x", md5.Sum(url))
	if short {
		return hash[:6]
	}
	return hash
}

func NewRouter() chi.Router {
	r := chi.NewRouter()
	dbm := DatabaseManager{}
	db := MapDatabase{db: &dbm}

	r.Route("/", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			inputUrl, err := io.ReadAll(r.Body)
			defer r.Body.Close()

			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if _, err := url.ParseRequestURI(string(inputUrl)); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			key := Normalize(HashURL(inputUrl, true))
			db.Create(key, string(inputUrl))

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
