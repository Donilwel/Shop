package test

import (
	"Shop/database/migrations"
	"Shop/database/models"
	"Shop/handlers"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

var testDB *gorm.DB

func TestMain(m *testing.M) {
	var err error

	dsn := fmt.Sprint(
		"host=postgres user=postgres password=123456 dbname=testdb port=5432 sslmode=disable",
	)

	testDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Не удалось подключиться к тестовой БД: " + err.Error())
	}

	migrations.DB.AutoMigrate(
		&models.User{},
		&models.Wallet{},
		&models.Transaction{},
	)

	exitCode := m.Run()
	os.Exit(exitCode)
}

// Очистка базы перед тестами
func clearDB() {
	testDB.Exec("TRUNCATE TABLE transactions, wallets, users RESTART IDENTITY CASCADE;")
}

// Создание тестового пользователя и кошелька
func createTestUser(username string, coins uint) (models.User, models.Wallet) {
	user := models.User{
		ID:       uuid.New(),
		Username: username,
	}
	testDB.Create(&user)

	wallet := models.Wallet{
		UserID: user.ID,
		Coin:   coins,
	}
	testDB.Create(&wallet)

	return user, wallet
}

// Мок Middleware для эмуляции аутентификации
func mockAuthMiddleware(userID uuid.UUID) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), "userID", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Тест: успешный перевод монет
func TestSendCoinHandler_Success(t *testing.T) {
	clearDB()

	sender, senderWallet := createTestUser("Sender", 100)
	receiver, receiverWallet := createTestUser("Receiver", 50)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"toUser": "Receiver",
		"coin":   30,
	})

	req := httptest.NewRequest(http.MethodPost, "/sendCoin", bytes.NewBuffer(reqBody))
	req = req.WithContext(context.WithValue(req.Context(), "userID", sender.ID))

	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	employeeSendCoinRouter := router.PathPrefix("/sendCoin").Subrouter()
	employeeSendCoinRouter.Use(mockAuthMiddleware(sender.ID))
	employeeSendCoinRouter.HandleFunc("", handlers.SendCoinHandler).Methods("POST")

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var updatedSenderWallet models.Wallet
	testDB.Where("user_id = ?", sender.ID).First(&updatedSenderWallet)
	assert.Equal(t, senderWallet.Coin-30, updatedSenderWallet.Coin)

	var updatedReceiverWallet models.Wallet
	testDB.Where("user_id = ?", receiver.ID).First(&updatedReceiverWallet)
	assert.Equal(t, receiverWallet.Coin+30, updatedReceiverWallet.Coin)
}

// Тест: попытка перевести самому себе
func TestSendCoinHandler_SelfTransfer(t *testing.T) {
	clearDB()

	sender, _ := createTestUser("Sender", 100)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"toUser": "Sender",
		"coin":   10,
	})

	req := httptest.NewRequest(http.MethodPost, "/sendCoin", bytes.NewBuffer(reqBody))
	req = req.WithContext(context.WithValue(req.Context(), "userID", sender.ID))

	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	employeeSendCoinRouter := router.PathPrefix("/sendCoin").Subrouter()
	employeeSendCoinRouter.Use(mockAuthMiddleware(sender.ID))
	employeeSendCoinRouter.HandleFunc("", handlers.SendCoinHandler).Methods("POST")

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
	assert.Contains(t, rr.Body.String(), "Самому себе нельзя перевести деньги")
}

// Тест: получатель не найден
func TestSendCoinHandler_ReceiverNotFound(t *testing.T) {
	clearDB()

	sender, _ := createTestUser("Sender", 100)

	reqBody, _ := json.Marshal(map[string]interface{}{
		"toUser": "UnknownUser",
		"coin":   20,
	})

	req := httptest.NewRequest(http.MethodPost, "/sendCoin", bytes.NewBuffer(reqBody))
	req = req.WithContext(context.WithValue(req.Context(), "userID", sender.ID))

	rr := httptest.NewRecorder()

	router := mux.NewRouter()
	employeeSendCoinRouter := router.PathPrefix("/sendCoin").Subrouter()
	employeeSendCoinRouter.Use(mockAuthMiddleware(sender.ID))
	employeeSendCoinRouter.HandleFunc("", handlers.SendCoinHandler).Methods("POST")

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
	assert.Contains(t, rr.Body.String(), "Получатель не найден")
}
