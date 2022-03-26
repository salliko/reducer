package handlers

import (
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/salliko/reducer/config"
	"github.com/salliko/reducer/internal/databases"
	"github.com/salliko/reducer/internal/datahashes"
	"io"
	"net/http"
	"net/url"
)

func InsertURL(URL []byte, hashURL datahashes.Hasing, db databases.Database, cfg config.Config, userID string) (string, error) {
	key := hashURL.Hash(URL)
	err := db.Create(key, string(URL), userID)
	if err != nil {
		if errors.Is(err, databases.ErrConflict) {
			return fmt.Sprintf("%s/%s", cfg.BaseURL, key), databases.ErrConflict
		} else {
			return "", err
		}
	}
	return fmt.Sprintf("%s/%s", cfg.BaseURL, key), nil
}

func GenerateShortURL(hashURL datahashes.Hasing, db databases.Database, cfg config.Config) http.HandlerFunc {
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
			if errors.Is(err, databases.ErrConflict) {
				w.WriteHeader(http.StatusConflict)
			} else {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			w.WriteHeader(http.StatusCreated)
		}

		w.Write([]byte(newURL))
	}
}

func RedirectFromShortToFull(db databases.Database) http.HandlerFunc {
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

func GenerateShortenJSONURL(hashURL datahashes.Hasing, db databases.Database, cfg config.Config) http.HandlerFunc {
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
			if errors.Is(err, databases.ErrConflict) {
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.WriteHeader(http.StatusConflict)
			} else {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusCreated)
		}

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

func GetAllShortenURLS(db databases.Database, cfg config.Config) http.HandlerFunc {
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

		allRows, err := db.SelectAll(cookie.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if len(allRows) == 0 {
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}

		for _, value := range allRows {
			shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, value.Hash)
			rows = append(rows, rowData{ShortURL: shortURL, OriginalURL: value.Original})
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

func Ping(db databases.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := db.Ping()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func GenerateManyShortenJSONURL(hashURL datahashes.Hasing, db databases.Database, cfg config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		var inputValues []databases.InputURL
		var outputValues []databases.OutputURL

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

		if err := json.NewDecoder(reader).Decode(&inputValues); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cookie, err := r.Cookie("user_id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		for _, value := range inputValues {
			key := hashURL.Hash([]byte(value.OriginalURL))
			err := db.CreateMany(databases.URL{
				Hash:     key,
				Original: value.OriginalURL,
				UserID:   cookie.Value,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
			outputValues = append(outputValues, databases.OutputURL{
				ShortURL:      fmt.Sprintf("%s/%s", cfg.BaseURL, hashURL.Hash([]byte(value.OriginalURL))),
				CorrelationID: value.CorrelationID,
			})
		}

		err = db.Flush()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusCreated)

		data, err := json.Marshal(outputValues)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Write(data)
	}
}
