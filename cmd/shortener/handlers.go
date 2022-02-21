package main

import (
	"encoding/json"
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

func GenerateShortenJSONURL(hashURL Hasing, db *MapDatabase) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var v struct {
			Url string `json:"url"`
		}

		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if _, err := url.ParseRequestURI(v.Url); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		key := hashURL.Hash([]byte(v.Url))
		db.Create(key, v.Url)

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Header().Add("Accept", "application/json")
		w.WriteHeader(http.StatusCreated)
		newURL := fmt.Sprintf("%s/%s", fullHostPath, key)

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
