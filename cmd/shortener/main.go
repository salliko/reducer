package main

import (
	"crypto/md5"
	"fmt"
	"github.com/go-chi/chi"
	"io"
	"log"
	"net/http"
	"strings"
)

var DB = make(map[string]string)

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

	r.Route("/", func(r chi.Router) {
		r.Post("/", func(w http.ResponseWriter, r *http.Request) {
			b, err := io.ReadAll(r.Body)
			defer r.Body.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			body := string(b)
			key := Normalize(HashURL(b, true))
			DB[key] = body
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(key))
		})
		r.Get("/{ID}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "ID")
			if val, ok := DB[id]; ok {
				//w.Header().Set("Location", val)
				//w.WriteHeader(http.StatusTemporaryRedirect)
				http.Redirect(w, r, val, http.StatusTemporaryRedirect)
			} else {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Not found"))
			}
		})
	})

	return r
}

func main() {
	r := NewRouter()
	log.Fatal(http.ListenAndServe(":8080", r))
}