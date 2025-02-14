package test

import (
	"Shop/database/migrations"
	"Shop/database/models"
	"Shop/handlers"
	"bytes"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
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
	migrations.InitDB()
	os.Exit(m.Run())
}

func setupTestDB() {
	if migrations.DB == nil {
		log.Fatal("Database connection is not initialized")
	}
	migrations.DB.Exec("DELETE FROM users")
}

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
