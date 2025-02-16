package handlers_test

import (
	"Shop/utils"
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSONFormat_Success(t *testing.T) {
	type Response struct {
		Message string `json:"message"`
	}

	response := Response{Message: "Привет, мир!"}

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Не удалось создать запрос: %v", err)
	}

	rr := httptest.NewRecorder()

	utils.JSONFormat(rr, req, response)

	assert.Equal(t, http.StatusOK, rr.Code)

	var result map[string]string
	err = json.NewDecoder(rr.Body).Decode(&result)
	if err != nil {
		t.Fatalf("Не удалось распарсить JSON: %v", err)
	}

	assert.Equal(t, "Привет, мир!", result["message"])
}

func TestJSONFormat_ErrorOnMarshal(t *testing.T) {
	response := make(chan int)

	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Fatalf("Не удалось создать запрос: %v", err)
	}

	rr := httptest.NewRecorder()

	utils.JSONFormat(rr, req, response)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)

	body := rr.Body.String()
	assert.Contains(t, body, "Ошибка сервера")
}
