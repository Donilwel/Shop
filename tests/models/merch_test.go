package models_test

import (
	"Shop/config"
	"Shop/database/migrations"
	"Shop/database/models"
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

func TestCreateMerch(t *testing.T) {
	SetupTestDB()

	merch := models.Merch{
		Name:  "TestMerch",
		Price: 100,
	}

	result := migrations.DB.Create(&merch)

	assert.NoError(t, result.Error)
	assert.NotEqual(t, uuid.Nil, merch.ID)
	assert.Equal(t, "TestMerch", merch.Name)
	assert.Equal(t, uint(100), merch.Price)
}

func TestCreateMerchWithDuplicateName(t *testing.T) {
	SetupTestDB()

	merch1 := models.Merch{
		Name:  "TestMerch",
		Price: 100,
	}
	migrations.DB.Create(&merch1)

	merch2 := models.Merch{
		Name:  "TestMerch",
		Price: 200,
	}

	result := migrations.DB.Create(&merch2)

	assert.Error(t, result.Error)
}
