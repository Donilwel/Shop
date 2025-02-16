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

func TestShowEmployeesHandler_Success(t *testing.T) {
	setupTestDB()
	userID := uuid.New()

	employee := models.User{
		ID:       uuid.New(),
		Username: "TestEmployee",
		Email:    "employee@example.com",
		Role:     models.EMPLOYEE_ROLE,
	}
	migrations.DB.Create(&employee)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	ctx := req.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.ShowEmployeesHandler(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestShowEmployeesHandler_NotFound(t *testing.T) {
	setupTestDB()
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	ctx := req.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.ShowEmployeesHandler(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestInformationHandler_Success(t *testing.T) {
	setupTestDB()
	userID := uuid.New()

	user := models.User{
		ID:       userID,
		Username: "TestUser",
		Email:    "testuser@example.com",
		Role:     models.EMPLOYEE_ROLE,
	}
	migrations.DB.Create(&user)

	wallet := models.Wallet{
		UserID: userID,
	}
	migrations.DB.Create(&wallet)

	req := httptest.NewRequest(http.MethodGet, "/info", nil)
	ctx := req.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.InformationHandler(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var response struct {
		Coins int `json:"coins"`
	}
	err := json.NewDecoder(res.Body).Decode(&response)
	assert.NoError(t, err, "Ошибка декодирования JSON")
	assert.Equal(t, 1000, response.Coins, "Неверное количество монет")
}

func TestInformationHandler_PartialData(t *testing.T) {
	setupTestDB()
	userID := uuid.New()

	user := models.User{
		ID:       userID,
		Username: "TestUser",
		Email:    "testuser@example.com",
	}
	migrations.DB.Create(&user)

	wallet := models.Wallet{
		UserID: userID,
	}
	migrations.DB.Create(&wallet)

	migrations.DB.Exec("INSERT INTO transactions (from_user, to_user, amount) VALUES (?, ?, ?)", uuid.New(), uuid.New(), 50)

	req := httptest.NewRequest(http.MethodGet, "/info", nil)
	ctx := req.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.InformationHandler(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestSendCoinHandler_Success(t *testing.T) {
	setupTestDB()

	senderID := uuid.New()
	receiver := models.User{ID: uuid.New(), Username: "receiver", Email: "receiver@example.com"}
	sender := models.User{ID: senderID, Username: "sender", Email: "sender@example.com"}

	migrations.DB.Create(&sender)
	migrations.DB.Create(&receiver)
	migrations.DB.Create(&models.Wallet{UserID: sender.ID, Coin: 100})
	migrations.DB.Create(&models.Wallet{UserID: receiver.ID, Coin: 50})

	requestBody, _ := json.Marshal(map[string]interface{}{
		"toUser": "receiver",
		"coin":   10,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewReader(requestBody))
	ctx := req.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, senderID)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.SendCoinHandler(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var transaction models.Transaction
	migrations.DB.First(&transaction)
	assert.Equal(t, sender.ID, transaction.FromUser)
	assert.Equal(t, receiver.ID, transaction.ToUser)
	assert.Equal(t, uint(10), transaction.Amount)
}

func TestSendCoinHandler_NotEnoughCoins(t *testing.T) {
	setupTestDB()

	senderID := uuid.New()
	receiver := models.User{ID: uuid.New(), Username: "receiver", Email: "receiver@example.com"}
	sender := models.User{ID: senderID, Username: "sender", Email: "sender@example.com"}

	migrations.DB.Create(&sender)
	migrations.DB.Create(&receiver)
	migrations.DB.Create(&models.Wallet{UserID: sender.ID, Coin: 5})
	migrations.DB.Create(&models.Wallet{UserID: receiver.ID, Coin: 50})

	requestBody, _ := json.Marshal(map[string]interface{}{
		"toUser": "receiver",
		"coin":   10,
	})

	r := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewReader(requestBody))
	ctx := r.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, senderID)
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.SendCoinHandler(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Недостаточно монет на балансе")
}

func TestSendCoinHandler_RecipientNotFound(t *testing.T) {
	setupTestDB()

	senderID := uuid.New()
	sender := models.User{ID: senderID, Username: "sender"}

	migrations.DB.Create(&sender)
	migrations.DB.Create(&models.Wallet{UserID: sender.ID, Coin: 100})

	requestBody, _ := json.Marshal(map[string]interface{}{
		"toUser": "nonexistent",
		"coin":   10,
	})

	r := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewReader(requestBody))
	ctx := r.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, senderID)
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.SendCoinHandler(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Получатель не найден")
}

func TestSendCoinHandler_SenderNotFound(t *testing.T) {
	setupTestDB()

	receiver := models.User{ID: uuid.New(), Username: "receiver"}
	migrations.DB.Create(&receiver)
	migrations.DB.Create(&models.Wallet{UserID: receiver.ID, Coin: 50})

	senderID := uuid.New()
	requestBody, _ := json.Marshal(map[string]interface{}{
		"toUser": "receiver",
		"coin":   10,
	})

	r := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewReader(requestBody))
	ctx := r.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, senderID)
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.SendCoinHandler(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Кошелек отправителя не найден")
}

func TestSendCoinHandler_JWTNotFound(t *testing.T) {
	setupTestDB()

	senderID := uuid.New()
	receiver := models.User{ID: uuid.New(), Username: "receiver", Email: "receiver@example.com"}
	sender := models.User{ID: senderID, Username: "sender", Email: "sender@example.com"}

	migrations.DB.Create(&sender)
	migrations.DB.Create(&receiver)
	migrations.DB.Create(&models.Wallet{UserID: sender.ID, Coin: 100})
	migrations.DB.Create(&models.Wallet{UserID: receiver.ID, Coin: 50})

	requestBody, _ := json.Marshal(map[string]interface{}{
		"toUser": "receiver",
		"coin":   10,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewReader(requestBody))
	ctx := req.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, uuid.New())
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handlers.SendCoinHandler(w, req)
	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestSendCoinHandler_SendYourself(t *testing.T) {
	setupTestDB()

	senderID := uuid.New()
	receiver := models.User{ID: uuid.New(), Username: "receiver", Email: "receiver@example.com"}
	sender := models.User{ID: senderID, Username: "sender", Email: "sender@example.com"}

	migrations.DB.Create(&sender)
	migrations.DB.Create(&receiver)
	migrations.DB.Create(&models.Wallet{UserID: sender.ID, Coin: 100})
	migrations.DB.Create(&models.Wallet{UserID: receiver.ID, Coin: 50})

	requestBody, _ := json.Marshal(map[string]interface{}{
		"toUser": "sender",
		"coin":   10,
	})

	req := httptest.NewRequest(http.MethodPost, "/api/sendCoin", bytes.NewReader(requestBody))
	ctx := req.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, senderID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	handlers.SendCoinHandler(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Самому себе нельзя перевести деньги")
}

func TestBuyItemHandler_UserNotFound(t *testing.T) {
	setupTestDB()

	userID := uuid.New()
	itemName := "TestItem"

	r := httptest.NewRequest(http.MethodGet, "/buy/"+itemName, nil)
	ctx := r.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, userID)
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.BuyItemHandler(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Покупатель не найден в базе данных")
}

func TestBuyItemHandler_ItemNotFound(t *testing.T) {
	setupTestDB()

	user := models.User{ID: uuid.New(), Username: "buyer", Email: "byuer@example.com"}
	migrations.DB.Create(&user)

	wallet := models.Wallet{UserID: user.ID, Coin: 100}
	migrations.DB.Create(&wallet)

	userID := user.ID
	itemName := "NonExistentItem"

	r := httptest.NewRequest(http.MethodGet, "/api/buy/"+itemName, nil)
	ctx := r.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, userID)
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.BuyItemHandler(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Запрошенная вещь не существует в базе данных")
}

func TestGetSource(t *testing.T) {
	t.Run("When fromCache is true", func(t *testing.T) {
		result := handlers.GetSource(true)
		assert.Equal(t, "Redis", result)
	})

	t.Run("When fromCache is false", func(t *testing.T) {
		result := handlers.GetSource(false)
		assert.Equal(t, "PostgreSQL", result)
	})
}

func TestBuyItemHandler_WalletNotFound(t *testing.T) {
	setupTestDB()

	merch := models.Merch{Name: "TestMerch", Price: 100}
	migrations.DB.Create(&merch)

	user := models.User{ID: uuid.New(), Username: "buyer", Email: "buyer@example.com"}
	migrations.DB.Create(&user)

	userID := user.ID

	r := httptest.NewRequest(http.MethodGet, "/api/buy/"+merch.Name, nil)
	ctx := r.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, userID)
	r = r.WithContext(ctx)
	w := httptest.NewRecorder()

	handlers.BuyItemHandler(w, r)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "Кошелька покупателя не существует в базе данных")
}

//func TestBuyItemHandler_InsufficientFunds(t *testing.T) {
//	setupTestDB()
//
//	user := models.User{ID: uuid.New(), Username: "buyer"}
//	migrations.DB.Create(&user)
//	merch := models.Merch{Name: "TestItem", Price: 50}
//	migrations.DB.Create(&merch)
//	wallet := models.Wallet{UserID: user.ID, Coin: 10}
//	migrations.DB.Create(&wallet)
//
//	r := httptest.NewRequest(http.MethodGet, "/api/buy/"+merch.Name, nil)
//	ctx := r.Context()
//	ctx = context.WithValue(ctx, utils.UserIDKey, user.ID)
//	r = r.WithContext(ctx)
//	w := httptest.NewRecorder()
//
//	handlers.BuyItemHandler(w, r)
//
//	assert.Equal(t, http.StatusBadRequest, w.Code)
//	assert.Contains(t, w.Body.String(), "Недостаточно средств на кошельке у пользователя.")
//}
//
//func TestBuyItemHandler_Success(t *testing.T) {
//	setupTestDB()
//
//	// Создаем мерч
//	merch := models.Merch{Name: "TestItem", Price: 50}
//	err := migrations.DB.Create(&merch).Error
//	if err != nil {
//		t.Fatalf("Ошибка при создании мерча: %v", err)
//	}
//
//	// Создаем пользователя
//	user := models.User{ID: uuid.New(), Username: "buyer", Email: "buyer@example.com"}
//	err = migrations.DB.Create(&user).Error
//	if err != nil {
//		t.Fatalf("Ошибка при создании пользователя: %v", err)
//	}
//
//	// Создаем кошелек для пользователя
//	wallet := models.Wallet{UserID: user.ID, Coin: 100}
//	err = migrations.DB.Create(&wallet).Error
//	if err != nil {
//		t.Fatalf("Ошибка при создании кошелька: %v", err)
//	}
//
//	// Проверяем, что данные действительно созданы
//	var createdMerch models.Merch
//	err = migrations.DB.First(&createdMerch, "name = ?", "TestItem").Error
//	if err != nil {
//		t.Fatalf("Ошибка при проверке существования мерча: %v", err)
//	}
//
//	var createdWallet models.Wallet
//	err = migrations.DB.First(&createdWallet, "user_id = ?", user.ID).Error
//	if err != nil {
//		t.Fatalf("Ошибка при проверке существования кошелька: %v", err)
//	}
//
//	// Запрос для покупки
//	r := httptest.NewRequest(http.MethodGet, "/api/buy/TestItem", nil)
//	ctx := r.Context()
//	ctx = context.WithValue(ctx, utils.UserIDKey, user.ID)
//	r = r.WithContext(ctx)
//	w := httptest.NewRecorder()
//
//	// Выполнение обработчика
//	handlers.BuyItemHandler(w, r)
//
//	// Проверяем ответ
//	assert.Equal(t, http.StatusOK, w.Code)
//
//	// Проверяем, что баланс уменьшился
//	var updatedWallet models.Wallet
//	err = migrations.DB.First(&updatedWallet, "user_id = ?", user.ID).Error
//	if err != nil {
//		t.Fatalf("Ошибка при получении обновленного кошелька: %v", err)
//	}
//
//	// Баланс должен уменьшиться на цену товара
//	assert.Equal(t, uint(0x64), updatedWallet.Coin)
//}
