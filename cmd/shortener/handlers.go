package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/go-chi/chi"
)

func InsertURL(URL []byte, hashURL Hasing, db Database, cfg Config) (string, error) {
	key := hashURL.Hash(URL)
	err := db.Create(key, string(URL))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", cfg.BaseURL, key), nil
}

func GenerateShortURL(hashURL Hasing, db Database, cfg Config) http.HandlerFunc {
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

		newURL, err := InsertURL(inputURL, hashURL, db, cfg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(newURL))
	}
}

func RedirectFromShortToFull(db Database) http.HandlerFunc {
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

func GenerateShortenJSONURL(hashURL Hasing, db Database, cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var v struct {
			URL string `json:"url"`
		}

		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if _, err := url.ParseRequestURI(v.URL); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		newURL, err := InsertURL([]byte(v.URL), hashURL, db, cfg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusCreated)

		res := struct {
			Result string `json:"result"`
		}{
			Result: newURL,
		}

		data, err := json.Marshal(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Write(data)
	}
}
