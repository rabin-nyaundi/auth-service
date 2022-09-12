package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ping(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

func TestPing(t *testing.T) {
	rr := httptest.NewRecorder()

	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	ping(rr, r)

	rs := rr.Result()

	assert.Equal(t, rs.StatusCode, http.StatusAccepted)
}
