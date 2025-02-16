package handlers_test

import (
	"Shop/database/migrations"
	"Shop/database/models"
	"Shop/handlers"
	"Shop/utils"
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestShowMerchHandler_Success(t *testing.T) {
	SetupTestDB()
	userID := uuid.New()
	merch := models.Merch{Name: "TestMerch", Price: 100}
	migrations.DB.Create(&merch)

	req := httptest.NewRequest(http.MethodGet, "/merch", nil)
	ctx := req.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.ShowMerchHandler(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Contains(t, w.Body.String(), "TestMerch")
}

func TestShowMerchHandler_NotFound(t *testing.T) {
	SetupTestDB()
	userID := uuid.New()

	req := httptest.NewRequest(http.MethodGet, "/merch", nil)
	ctx := req.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.ShowMerchHandler(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestShowMerchHandler_Timeout(t *testing.T) {
	SetupTestDB()
	userID := uuid.New()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	merch := models.Merch{Name: "TestMerch", Price: 100}
	migrations.DB.Create(&merch)

	req := httptest.NewRequest(http.MethodGet, "/merch", nil)
	req = req.WithContext(ctx)
	ctx = req.Context()
	ctx = context.WithValue(ctx, utils.UserIDKey, userID)
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()
	handlers.ShowMerchHandler(w, req)
	res := w.Result()
	defer res.Body.Close()

	assert.Equal(t, http.StatusRequestTimeout, res.StatusCode)
	assert.Contains(t, w.Body.String(), "Запрос отменен клиентом")
}
