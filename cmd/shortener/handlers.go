package main

import (
	"fmt"
	"github.com/go-chi/chi"
	"io"
	"net/http"
	"net/url"
)

func GenerateShortURL(hashURL Hasing, db *MapDatabase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inputURL, err := io.ReadAll(r.Body)
		defer r.Body.Close()

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := url.ParseRequestURI(string(inputURL)); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		key := hashURL.Hash(inputURL)
		db.Create(key, string(inputURL))

		w.WriteHeader(http.StatusCreated)
		newURL := fmt.Sprintf("%s/%s", fullHostPath, key)
		w.Write([]byte(newURL))
	}
}

func RedirectFromShortToFull(db *MapDatabase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "ID")
		val, err := db.Select(id)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Not found"))
			return
		}
		http.Redirect(w, r, val, http.StatusTemporaryRedirect)
		w.Write([]byte("Found"))
	}
}
