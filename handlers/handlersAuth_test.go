package handlers_test

import (
	"Shop/config"
	"Shop/database/migrations"
	"Shop/database/models"
	"Shop/handlers"
	"Shop/utils"
	"bytes"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthHandler_SuccessfulLogin(t *testing.T) {
	setupTestDB()
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := models.User{
		ID:       uuid.New(),
		Username: "TestUser",
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}
	migrations.DB.Create(&user)

	requestBody, _ := json.Marshal(map[string]string{
		"email":    user.Email,
		"password": password,
	})

	req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewReader(requestBody))
	w := httptest.NewRecorder()

	handlers.AuthHandler(w, req)
	res := w.Result()
	defer res.Body.Close()
	err := migrations.DB.Where("email = ?", "test@example.com").First(&user).Error
	assert.NoError(t, err, "Пользователь должен быть создан в БД")

	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestAuthHandler_InvalidPassword(t *testing.T) {
	setupTestDB()
	password := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	user := models.User{
		ID:       uuid.New(),
		Username: "TestUser",
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}
	migrations.DB.Create(&user)

	requestBody, _ := json.Marshal(map[string]string{
		"email":    user.Email,
		"password": "wrongpassword",
	})

	req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewReader(requestBody))
	w := httptest.NewRecorder()

	handlers.AuthHandler(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

type AuthRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"securepassword"`
}

func TestAuthHandler_CreateNewUser(t *testing.T) {
	setupTestDB()
	requestBody, _ := json.Marshal(map[string]string{
		"email":    "newuser@example.com",
		"password": "newpassword",
	})

	req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewReader(requestBody))
	w := httptest.NewRecorder()

	handlers.AuthHandler(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusCreated, res.StatusCode)

	var user models.User
	err := migrations.DB.Where("email = ?", "newuser@example.com").First(&user).Error
	assert.NoError(t, err, "Пользователь должен быть создан в БД")
}

func TestAuthHandler_InvalidJSON(t *testing.T) {
	setupTestDB()
	req := httptest.NewRequest(http.MethodPost, "/auth", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handlers.AuthHandler(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusBadRequest, res.StatusCode)
}

func TestLogoutHandler_SuccessfulLogout(t *testing.T) {
	setupTestDB()
	userID := uuid.New()
	token := "valid_token"
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	ctx := req.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.LogoutHandler(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)

	var revokedToken models.RevokedToken
	err := migrations.DB.Where("token = ?", token).First(&revokedToken).Error
	assert.NoError(t, err, "Токен должен быть сохранен в таблице revoked_tokens")

	exists, err := config.Rdb.Exists(context.Background(), token).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), exists, "Токен должен быть удален из Redis")
}

func TestLogoutHandler_MissingToken(t *testing.T) {
	setupTestDB()
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	w := httptest.NewRecorder()

	handlers.LogoutHandler(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}

func TestLogoutHandler_InvalidTokenFormat(t *testing.T) {
	setupTestDB()
	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "InvalidTokenFormat")
	w := httptest.NewRecorder()

	handlers.LogoutHandler(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode)
}
