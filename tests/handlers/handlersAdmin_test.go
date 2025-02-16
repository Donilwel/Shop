package handlers_test

import (
	"Shop/database/migrations"
	"Shop/database/models"
	"Shop/handlers"
	"Shop/utils"
	"bytes"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPutMoneyHandler_Success(t *testing.T) {
	SetupTestDB()

	// Создание пользователя и кошелька
	user := models.User{ID: uuid.New(), Username: "admin", Email: "admin@example.com"}
	migrations.DB.Create(&user)
	wallet := models.Wallet{UserID: user.ID, Coin: 1000}
	migrations.DB.Create(&wallet)

	worker := models.User{ID: uuid.New(), Username: "worker", Email: "worker@example.com"}
	migrations.DB.Create(&worker)
	workerWallet := models.Wallet{UserID: worker.ID, Coin: 500}
	migrations.DB.Create(&workerWallet)

	requestBody := handlers.SendMoney{
		NickTaker: "worker",
		Coin:      100,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Не удалось сериализовать тело запроса: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	ctx := req.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.PutMoneyHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Перевод монет успешен")

	var updatedWallet models.Wallet
	migrations.DB.First(&updatedWallet, "user_id = ?", worker.ID)
	assert.Equal(t, uint(600), updatedWallet.Coin)
}

func TestPutMoneyHandler_InvalidRequestBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/admin/users", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	w := httptest.NewRecorder()
	handlers.PutMoneyHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Некорректное тело запроса")
}

func TestPutMoneyHandler_InvalidCoinAmount(t *testing.T) {
	SetupTestDB()

	user := models.User{ID: uuid.New(), Username: "admin", Email: "admin@example.com"}
	migrations.DB.Create(&user)
	wallet := models.Wallet{UserID: user.ID, Coin: 1000}
	migrations.DB.Create(&wallet)

	requestBody := handlers.SendMoney{
		NickTaker: "worker",
		Coin:      2000,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Не удалось сериализовать тело запроса: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	ctx := req.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.PutMoneyHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Количество монет должно быть в диапазоне от 1 до 1000 включительно")
}

func TestPutMoneyHandler_UserNotFound(t *testing.T) {
	SetupTestDB()

	user := models.User{ID: uuid.New(), Username: "admin", Email: "admin@example.com"}
	migrations.DB.Create(&user)
	wallet := models.Wallet{UserID: user.ID, Coin: 1000}
	migrations.DB.Create(&wallet)

	requestBody := handlers.SendMoney{
		NickTaker: "nonexistent",
		Coin:      100,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Не удалось сериализовать тело запроса: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	ctx := req.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.PutMoneyHandler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Работник с никнеймом nonexistent не найден.")
}

func TestPutMoneyHandler_WalletNotFound(t *testing.T) {
	SetupTestDB()

	user := models.User{ID: uuid.New(), Username: "admin", Email: "admin@example.com"}
	migrations.DB.Create(&user)
	wallet := models.Wallet{UserID: user.ID, Coin: 1000}
	migrations.DB.Create(&wallet)

	worker := models.User{ID: uuid.New(), Username: "worker", Email: "worker@example.com"}
	migrations.DB.Create(&worker)

	requestBody := handlers.SendMoney{
		NickTaker: "worker",
		Coin:      100,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Не удалось сериализовать тело запроса: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	ctx := req.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.PutMoneyHandler(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Кошелек получателя не найден.")
}

func TestAddOrChangeMerchHandler_Success_Add(t *testing.T) {
	SetupTestDB()

	user := models.User{ID: uuid.New(), Username: "admin", Email: "admin@example.com"}
	migrations.DB.Create(&user)

	requestBody := handlers.MerchInfo{
		Type:  "NewMerch",
		Price: 500,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Не удалось сериализовать тело запроса: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/merch", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	ctx := req.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.AddOrChangeMerchHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Был создан новый мерч: NewMerch")

	var merch models.Merch
	migrations.DB.First(&merch, "name = ?", "NewMerch")
	assert.Equal(t, uint(500), merch.Price)
}

func TestAddOrChangeMerchHandler_Success_Update(t *testing.T) {
	SetupTestDB()

	user := models.User{ID: uuid.New(), Username: "admin", Email: "admin@example.com"}
	migrations.DB.Create(&user)

	merch := models.Merch{Name: "ExistingMerch", Price: 300}
	migrations.DB.Create(&merch)

	requestBody := handlers.MerchInfo{
		Type:  "ExistingMerch",
		Price: 500,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Не удалось сериализовать тело запроса: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/merch", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	ctx := req.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.AddOrChangeMerchHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Цена мерча ExistingMerch была обновлена")

	var updatedMerch models.Merch
	migrations.DB.First(&updatedMerch, "name = ?", "ExistingMerch")
	assert.Equal(t, uint(500), updatedMerch.Price)
}

func TestAddOrChangeMerchHandler_InvalidRequestBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/merch", nil)
	req.Header.Set("Authorization", "Bearer valid-token")

	w := httptest.NewRecorder()
	handlers.AddOrChangeMerchHandler(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Не удалось получить userID")
}

func TestAddOrChangeMerchHandler_InvalidPrice(t *testing.T) {
	SetupTestDB()

	user := models.User{ID: uuid.New(), Username: "admin", Email: "admin@example.com"}
	migrations.DB.Create(&user)

	requestBody := handlers.MerchInfo{
		Type:  "NewMerch",
		Price: 1500,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Не удалось сериализовать тело запроса: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/merch", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	ctx := req.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.AddOrChangeMerchHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Цена мерча должна быть в диапазоне от 1 до 1000 включительно")
}

func TestAddOrChangeMerchHandler_TypeEmpty(t *testing.T) {
	SetupTestDB()

	user := models.User{ID: uuid.New(), Username: "admin", Email: "admin@example.com"}
	migrations.DB.Create(&user)

	requestBody := handlers.MerchInfo{
		Type:  "",
		Price: 500,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Не удалось сериализовать тело запроса: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/merch", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	ctx := req.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.AddOrChangeMerchHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Тип мерча не должен быть пустым")
}

func TestAddOrChangeMerchHandler_MerchAlreadyExists(t *testing.T) {
	SetupTestDB()

	user := models.User{ID: uuid.New(), Username: "admin", Email: "admin@example.com"}
	migrations.DB.Create(&user)

	merch := models.Merch{Name: "ExistingMerch", Price: 300}
	migrations.DB.Create(&merch)

	requestBody := handlers.MerchInfo{
		Type:  "ExistingMerch",
		Price: 300,
	}
	body, err := json.Marshal(requestBody)
	if err != nil {
		t.Fatalf("Не удалось сериализовать тело запроса: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/merch", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer valid-token")
	ctx := req.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, user.ID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.AddOrChangeMerchHandler(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Цена мерча совпадает с заданной")
}
