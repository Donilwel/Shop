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
