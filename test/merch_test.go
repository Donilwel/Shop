package test

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
)

func TestShowMerchHandlerSuccess(t *testing.T) {
	setupTestDB()
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
}

func TestShowMerchHandlerNotFound(t *testing.T) {
	setupTestDB()
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
