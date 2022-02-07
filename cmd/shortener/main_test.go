package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestIndexHandle(t *testing.T) {
	request := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("http://ya.ru"))
	w := httptest.NewRecorder()
	h := http.HandlerFunc(IndexHandle)
	h.ServeHTTP(w, request)
	res := w.Result()
	defer res.Body.Close()

	if res.StatusCode != 201 {
		t.Errorf("Expected status code %d, got 201", res.StatusCode)
	}
}
