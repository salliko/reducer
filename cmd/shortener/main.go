package main

import (
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"net/http"
)

var DB = make(map[string]string)

func HashURL(url []byte, short bool) string {
	hash := fmt.Sprintf("%x", md5.Sum(url))
	if short {
		return hash[:6]
	}
	return hash
}

func IndexHandle(rw http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		id := r.URL.Path[1:]
		if val, ok := DB[id]; ok {
			rw.Header().Set("Location", val)
			rw.WriteHeader(http.StatusTemporaryRedirect)
			//http.Redirect(w, r, val, http.StatusTemporaryRedirect)
		} else {
			rw.WriteHeader(http.StatusBadRequest)
			rw.Write([]byte("Not found"))
		}
	case http.MethodPost:
		b, err := io.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			http.Error(rw, err.Error(), http.StatusInternalServerError)
		}
		body := string(b)
		key := HashURL(b, true)
		DB[key] = body
		rw.WriteHeader(http.StatusCreated)
		rw.Write([]byte(key))
	default:
		http.Error(rw, "Wrong", http.StatusNotFound)
		return
	}
}

func main() {
	http.HandleFunc("/", IndexHandle)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
