package E2E

import (
	"Shop/config"
	"Shop/database/migrations"
	"context"
	"log"
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

func SetupTestDB() {
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
