package main

import (
	"compress/gzip"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi"
	_ "github.com/jackc/pgx/v4"
	"io"
	"net/http"
	"net/url"
)

func InsertURL(URL []byte, hashURL Hasing, db Database, cfg Config, userID string) (string, error) {
	key := hashURL.Hash(URL)
	err := db.Create(key, string(URL), userID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s", cfg.BaseURL, key), nil
}

func GenerateShortURL(hashURL Hasing, db Database, cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var reader io.Reader

		if r.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			reader = gz
			defer gz.Close()
		} else {
			reader = r.Body
		}

		inputURL, err := io.ReadAll(reader)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if _, err := url.ParseRequestURI(string(inputURL)); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cookie, err := r.Cookie("user_id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		newURL, err := InsertURL(inputURL, hashURL, db, cfg, cookie.Value)
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

		var reader io.Reader

		if r.Header.Get(`Content-Encoding`) == `gzip` {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			reader = gz
			defer gz.Close()
		} else {
			reader = r.Body
		}

		if err := json.NewDecoder(reader).Decode(&v); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if _, err := url.ParseRequestURI(v.URL); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cookie, err := r.Cookie("user_id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		newURL, err := InsertURL([]byte(v.URL), hashURL, db, cfg, cookie.Value)
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

func GetAllShortenURLS(db Database, cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type rowData struct {
			ShortURL    string `json:"short_url"`
			OriginalURL string `json:"original_url"`
		}

		var rows []rowData

		cookie, err := r.Cookie("user_id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		allRows := db.SelectAll(cookie.Value)
		if len(allRows) == 0 {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}

		for key, value := range allRows {
			shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, key)
			rows = append(rows, rowData{ShortURL: shortURL, OriginalURL: value})
		}

		data, err := json.Marshal(rows)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.Write(data)
	}
}

func Ping(cfg Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		db, err := sql.Open("postgres", cfg.DatabaseDSN)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer db.Close()

		//err = db.Ping()
		//
		//if err != nil {
		//	http.Error(w, err.Error(), http.StatusInternalServerError)
		//	return
		//}

		w.WriteHeader(http.StatusOK)
	}
}
