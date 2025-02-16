package handlers

import (
	"Shop/loging"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

// PingHandler отвечает на запрос "ping" строкой "pong".
//
// @Summary Проверка работоспособности сервера
// @Description Возвращает "pong", если сервер работает корректно
// @Tags Ping
// @Accept  json
// @Produce  plain
// @Success 200 {string} string "pong"
// @Failure 400 {string} string "Ошибка написания pong. Ошибка подключения"
// @Router /api/ping [get]
func PingHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("pong")); err != nil {
		loging.LogRequest(logrus.WarnLevel, uuid.Nil, r, http.StatusBadRequest, nil, startTime, "Ошибка написания pong. Ошибка подключения")
		return
	}
	loging.LogRequest(logrus.InfoLevel, uuid.Nil, r, http.StatusOK, nil, startTime, "Pong ответ успешно отправлен")
}
