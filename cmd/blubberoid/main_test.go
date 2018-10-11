package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBlubberoid(t *testing.T) {
	rec := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/test", strings.NewReader(`---
    version: v3
    base: foo
    variants:
      test: {}`))

	blubberoid(rec, req)

	resp := rec.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, "text/plain", resp.Header.Get("Content-Type"))
	assert.Contains(t, string(body), "FROM foo")
	assert.Contains(t, string(body), `LABEL blubber.variant="test"`)
}
