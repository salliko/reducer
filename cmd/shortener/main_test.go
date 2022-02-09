package main

import (
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
	r := NewRouter()
	ts := httptest.NewServer(r)
	defer ts.Close()

	resp := testRequest(t, ts, "POST", "/", strings.NewReader("http://yandex.ru"))
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	//assert.Equal(t, "brand:renault", body)

	//resp, _ = testRequest(t, ts, "GET", "/"+body, nil)
	//assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
}
