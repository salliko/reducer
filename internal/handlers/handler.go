package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/salliko/reducer/config"
	"github.com/salliko/reducer/internal/databases"
	"github.com/salliko/reducer/internal/datahashes"
	"io"
	"log"
	"net/http"
	"net/url"
	"sync"
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

		inputURL, err := io.ReadAll(r.Body)

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
			if errors.Is(err, databases.ErrGone) {
				http.Error(w, err.Error(), http.StatusGone)
				return
			}
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

		if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
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
				log.Println(err.Error())
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
			log.Println(err.Error())
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Println(w.Header())

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
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		allRows, err := db.SelectAll(cookie.Value)
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if len(allRows) == 0 {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusNoContent)
			http.Error(w, "No Content", http.StatusNoContent)
			return
		}

		for _, value := range allRows {
			shortURL := fmt.Sprintf("%s/%s", cfg.BaseURL, value.Hash)
			rows = append(rows, rowData{ShortURL: shortURL, OriginalURL: value.Original})
		}

		data, err := json.Marshal(rows)
		if err != nil {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusBadRequest)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
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

		if err := json.NewDecoder(r.Body).Decode(&inputValues); err != nil {
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

type deleteItem struct {
	Key    string
	UserID string
}

func Delete(db databases.Database) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var keys []string

		if err := json.NewDecoder(r.Body).Decode(&keys); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		cookie, err := r.Cookie("user_id")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		inputCh := make(chan deleteItem)
		workersCount := 10

		go func() {
			for _, key := range keys {
				inputCh <- deleteItem{Key: key, UserID: cookie.Value}
			}
			close(inputCh)
		}()

		fanOutChs := fanOut(inputCh, workersCount)
		workerChs := make([]chan error, 0, workersCount)
		for _, fanOutCh := range fanOutChs {
			w := newWorker(db, fanOutCh)
			workerChs = append(workerChs, w)
		}

		// здесь fanIn
		for v := range fanIn(workerChs...) {
			log.Println(v)
		}

		w.WriteHeader(http.StatusAccepted)
	}
}

func fanOut(inputCh chan deleteItem, n int) []chan deleteItem {
	chs := make([]chan deleteItem, 0, n)
	for i := 0; i < n; i++ {
		ch := make(chan deleteItem)
		chs = append(chs, ch)
	}

	go func() {
		defer func(chs []chan deleteItem) {
			for _, ch := range chs {
				close(ch)
			}
		}(chs)

		for i := 0; ; i++ {
			if i == len(chs) {
				i = 0
			}

			item, ok := <-inputCh
			if !ok {
				return
			}

			ch := chs[i]
			ch <- item
		}
	}()

	return chs
}

func newWorker(db databases.Database, inputCh <-chan deleteItem) chan error {
	outCh := make(chan error)

	go func() {
		for item := range inputCh {
			err := db.Delete(item.Key, item.UserID)
			outCh <- err
		}

		close(outCh)
	}()

	return outCh
}

func fanIn(inputChs ...chan error) (chan error) {
	outCh := make(chan error)

	go func() {
		wg := &sync.WaitGroup{}

		for _, inputCh := range inputChs {
			wg.Add(1)

			go func(inputCh chan error) {
				defer wg.Done()
				for item := range inputCh {
					outCh <- item
				}
			}(inputCh)
		}

		wg.Wait()
		close(outCh)
	}()

	return outCh
}