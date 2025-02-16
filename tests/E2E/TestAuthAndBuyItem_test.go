package E2E

import (
	"Shop/config"
	"Shop/database/migrations"
	"Shop/database/models"
	"Shop/utils"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"log"
	"net/http"
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

func TestAuthAndBuyItem(t *testing.T) {
	// Шаг 1: Генерация нового уникального имени пользователя
	SetupTestDB()
	username := utils.GenerateUsername()
	password := "securepassword"
	email := fmt.Sprintf("%s@example.com", username)

	// Шаг 2: Создание нового пользователя через API (если он ещё не существует)
	authRequest := map[string]string{
		"email":    email,
		"password": password,
	}

	authBody, err := json.Marshal(authRequest)
	assert.NoError(t, err)

	// Подготовка запроса авторизации
	req, err := http.NewRequest("POST", "http://localhost:8080/api/auth", bytes.NewBuffer(authBody))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	// Выполнение запроса авторизации
	client := &http.Client{}
	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer resp.Body.Close()

	// Проверка успешной авторизации
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expected status code 200")

	// Извлекаем JWT токен из ответа
	var authResponse map[string]string
	err = json.NewDecoder(resp.Body).Decode(&authResponse)
	assert.NoError(t, err)
	token, exists := authResponse["token"]
	assert.True(t, exists, "Expected token in response")

	// Шаг 3: Проверка покупки товара
	item := "SharkT-Shirt" // Пример товара

	// Подготовка запроса на покупку товара
	reqBuy, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:8080/api/buy/%s", item), nil)
	assert.NoError(t, err)
	reqBuy.Header.Set("Authorization", "Bearer "+token)

	// Выполнение запроса на покупку товара
	respBuy, err := client.Do(reqBuy)
	assert.NoError(t, err)
	defer respBuy.Body.Close()

	// Проверка успешной покупки товара
	assert.Equal(t, http.StatusOK, respBuy.StatusCode, "Expected status code 200")

	// Шаг 4: Проверка данных о пользователе после покупки
	var user models.User
	err = migrations.DB.Where("email = ?", email).First(&user).Error
	assert.NoError(t, err)

	var wallet models.Wallet
	err = migrations.DB.Where("user_id = ?", user.ID).First(&wallet).Error
	assert.NoError(t, err)

	// Проверяем, что деньги были списаны с кошелька
	assert.True(t, wallet.Coin < 1000, "Expected wallet balance to decrease after purchase")
}
