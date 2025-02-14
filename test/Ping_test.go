package test

import (
	"Shop/handlers"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPingHandler(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	handlers.PingHandler(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body := make([]byte, 4)
	_, err := resp.Body.Read(body)
	assert.NoError(t, err)
	assert.Equal(t, "pong", string(body))
}
