package main

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func testRequest(t *testing.T, ts *httptest.Server, method, path string, body io.Reader) *http.Response {
	req, err := http.NewRequest(method, ts.URL+path, body)
	require.NoError(t, err)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	return resp
}

func TestRouter(t *testing.T) {
	var hashURL Hasing = &Md5HashData{}

	type want struct {
		status int
		body   string
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
				status: http.StatusInternalServerError,
				body:   "parse \"bfgbfgbsfg\": invalid URI for request\n",
			},
		},
		{
			name:   "#3 GET",
			method: http.MethodGet,
			path:   fmt.Sprintf("/%s", hashURL.Hash([]byte("http://ya.ru"))),
			want: want{
				status: http.StatusOK,
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
	}

	r := NewRouter()
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
		})
	}
}
