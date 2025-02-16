package E2E

import (
	"Shop/config"
	"Shop/database/migrations"
	"Shop/database/models"
	"Shop/utils"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"io"
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

type InfoMain struct {
	Coins     uint `json:"coins"`
	Inventory []struct {
		Type     string `json:"type"`
		Quantity int    `json:"quantity"`
	} `json:"inventory"`
	CoinHistory struct {
		Received []struct {
			FromUser string `json:"fromUser"`
			Amount   uint   `json:"amount"`
		} `json:"received"`
		Sent []struct {
			ToUser string `json:"toUser"`
			Amount uint   `json:"amount"`
		} `json:"sent"`
	} `json:"coinHistory"`
}

func TestInformationHandler(t *testing.T) {
	// Подготовка тестовой базы данных
	SetupTestDB()

	// Создание пользователя
	userData := models.User{
		Email:    "testuser@example.com",
		Password: "testpassword",
		Username: utils.GenerateUsername(),
	}
	err := migrations.DB.Create(&userData).Error
	if err != nil {
		t.Fatalf("Ошибка при создании пользователя: %v", err)
	}

	var user models.User
	err = migrations.DB.Where("email = ?", userData.Email).First(&user).Error
	if err != nil {
		t.Fatalf("Пользователь не найден: %v", err)
	}

	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		t.Fatalf("Ошибка при создании токена: %v", err)
	}

	wallet := models.Wallet{
		UserID: user.ID,
		Coin:   1000,
	}
	err = migrations.DB.Create(&wallet).Error
	if err != nil {
		t.Fatalf("Ошибка при создании кошелька: %v", err)
	}

	req, err := http.NewRequest("GET", "http://localhost:8080/api/info", nil)
	if err != nil {
		t.Fatalf("Ошибка при создании запроса: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	// Проверка статуса ответа
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Ошибка аутентификации, тело ответа: %s", string(body))
	}

	// Проверка структуры ответа
	var response InfoMain
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		t.Fatalf("Ошибка при декодировании ответа: %v", err)
	}

	// Проверяем, что данные возвращаются корректно
	assert.Equal(t, uint(1000), response.Coins)
	assert.NotNil(t, response.Inventory)
	assert.NotNil(t, response.CoinHistory)
}

func TestInformationHandler_InvalidToken(t *testing.T) {
	req, err := http.NewRequest("GET", "http://localhost:8080/api/info", nil)
	if err != nil {
		t.Fatalf("Ошибка при создании запроса: %v", err)
	}
	req.Header.Set("Authorization", "Bearer invalid-token")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("Ошибка при отправке запроса: %v", err)
	}
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "invalid token")
}
