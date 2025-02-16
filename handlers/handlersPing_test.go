package handlers_test

import (
	"Shop/config"
	"Shop/database/migrations"
	"Shop/handlers"
	"context"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	os.Setenv("POSTGRES_HOST", "localhost")
	os.Setenv("POSTGRES_USERNAME", "testuser")
	os.Setenv("POSTGRES_PASSWORD", "testpassword")
	os.Setenv("POSTGRES_DATABASE", "testdb")
	os.Setenv("POSTGRES_PORT", "5433")
	os.Setenv("REDIS_HOST", "localhost")
	os.Setenv("REDIS_PORT", "6379")
	migrations.InitDB()
	config.InitRedis()
	os.Exit(m.Run())
}

func setupTestDB() {
	if migrations.DB == nil {
		log.Fatal("Database connection is not initialized")
	}
	migrations.DB.Exec("DELETE FROM users")
	migrations.DB.Exec("DELETE FROM revoked_tokens")
	migrations.DB.Exec("DELETE FROM merches")
	migrations.DB.Exec("DELETE FROM wallets")
	migrations.DB.Exec("DELETE FROM purchases")
	migrations.DB.Exec("DELETE FROM transactions")
	if config.Rdb != nil {
		config.Rdb.FlushAll(context.Background())
	}
}

func TestPingHandler(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	handlers.PingHandler(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.Equal(t, "pong", string(body))
}
