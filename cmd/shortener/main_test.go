package main

import (
	"fmt"
	"github.com/salliko/reducer/config"
	"github.com/salliko/reducer/internal/databases"
	"github.com/salliko/reducer/internal/datahashes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) *http.Response {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	client := &http.Client{}
	client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	resp, err := client.Do(req)
	require.NoError(t, err)

	return resp
}

func TestRouter(t *testing.T) {
	var hashURL datahashes.Hasing = &datahashes.Md5HashData{}

	cfg := config.Config{
		ServerAddress: "localhost:8080",
		BaseURL:       "http://localhost:8080",
		//FileStoragePath: "C:/Users/snup4/Learn/reducer/test_bd.txt",
		//DatabaseDSN: "postgres://postgres:postgres@localhost:5432/postgres",
	}

	var db databases.Database
	var err error
	if cfg.DatabaseDSN != "" {
		db, err = databases.NewPostgresqlDatabase(cfg)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
	} else if cfg.FileStoragePath != "" {
		db, err = databases.NewFileDatabase(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
		defer db.Close()
	} else {
		db = databases.NewMapDatabase()
		defer db.Close()
	}

	type want struct {
		status   int
		location string
		body     string
		header   string
	}

	tests := []struct {
		name   string
		url    string
		method string
		path   string
		want   want
	}{
		{
			name:   "#1 POST",
			url:    "http://ya.ru",
			method: http.MethodPost,
			path:   "/",
			want: want{
				status: http.StatusCreated,
				body:   fmt.Sprintf("http://localhost:8080/%s", hashURL.Hash([]byte("http://ya.ru"))),
			},
		},
		{
			name:   "#2 POST",
			url:    "bfgbfgbsfg",
			method: http.MethodPost,
			path:   "/",
			want: want{
				status: http.StatusBadRequest,
				body:   "parse \"bfgbfgbsfg\": invalid URI for request\n",
			},
		},
		{
			name:   "#3 GET",
			method: http.MethodGet,
			path:   fmt.Sprintf("/%s", hashURL.Hash([]byte("http://ya.ru"))),
			want: want{
				status:   http.StatusTemporaryRedirect,
				location: "http://ya.ru",
			},
		},
		{
			name:   "#4 GET",
			method: http.MethodGet,
			path:   "/sdfsdfsdf",
			want: want{
				status: http.StatusBadRequest,
				body:   "Not found",
			},
		},
		{
			name:   "#5 POST API",
			url:    `{"url": "http://www.dns-shop.ru/"}`,
			method: http.MethodPost,
			path:   "/api/shorten",
			want: want{
				status: http.StatusCreated,
				body:   `{"result":"http://localhost:8080/8982ac"}`,
				header: `application/json; charset=UTF-8`,
			},
		},
		{
			name:   "#6 POST API MANY",
			url:    `[{"correlation_id": "xxx", "original_url": "https://krasnodar.beeline.ru/"},{"correlation_id": "xxx", "original_url": "https://jic-energy.ru/"},{"correlation_id": "xxx", "original_url": "https://habr.com/"}]`,
			method: http.MethodPost,
			path:   "/api/shorten/batch",
			want: want{
				status: http.StatusCreated,
				body:   `[{"correlation_id":"xxx","short_url":"http://localhost:8080/3617bf"},{"correlation_id":"xxx","short_url":"http://localhost:8080/419929"},{"correlation_id":"xxx","short_url":"http://localhost:8080/6c5b1c"}]`,
			},
		},
		{
			name:   "#7 POST",
			url:    "http://ya.ru",
			method: http.MethodPost,
			path:   "/",
			want: want{
				status: http.StatusConflict,
				body:   fmt.Sprintf("http://localhost:8080/%s", hashURL.Hash([]byte("http://ya.ru"))),
			},
		},
		{
			name:   "#8 POST API",
			url:    `{"url": "http://ya.ru"}`,
			method: http.MethodPost,
			path:   "/api/shorten",
			want: want{
				status: http.StatusConflict,
				body:   `{"result":"http://localhost:8080/1b556b"}`,
			},
		},
	}

	r := NewRouter(cfg, db)
	ts := httptest.NewServer(r)
	defer ts.Close()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := testRequest(t, ts, tt.method, tt.path, strings.NewReader(tt.url))
			defer resp.Body.Close()
			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Equal(t, tt.want.status, resp.StatusCode)
			if tt.want.body != "" {
				assert.Equal(t, tt.want.body, string(body))
			}
			if tt.want.location != "" {
				assert.Equal(t, tt.want.location, resp.Header.Get("Location"))
			}
			if tt.want.header != "" {
				assert.Equal(t, tt.want.header, resp.Header.Get("Content-Type"))
			}
		})
	}
}
