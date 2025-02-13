package handlers

import (
	"Shop/config"
	"Shop/database/migrations"
	"Shop/database/models"
	"Shop/loging"
	"Shop/utils"
	"context"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func ShowMerchHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	userID, _ := r.Context().Value(utils.UserIDKey).(uuid.UUID)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var merches []models.Merch
	cacheKey := "merch:all"

	select {
	case <-ctx.Done():
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusRequestTimeout, nil, startTime, "Запрос отменен клиентом")
		return
	default:
	}

	fromCache, err := utils.GetOrSetCache(ctx, config.Rdb, migrations.DB, cacheKey, migrations.DB, &merches, 5*time.Minute)
	if err != nil {
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка при поиске мерча.")
		http.Error(w, "Error fetching couriers", http.StatusInternalServerError)
		return
	}

	if len(merches) == 0 {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, nil, startTime, "Мерч не найден")
		http.Error(w, "No couriers found", http.StatusNotFound)
		return
	}
	data := "postgreSQL"
	if fromCache {
		data = "redis"
	}
	utils.JSONFormat(w, r, merches)
	loging.LogRequest(logrus.InfoLevel, userID, r, http.StatusOK, nil, startTime, "Список мерча показан успешно с помощью "+data)
}
