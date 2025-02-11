package handlers

import (
	"Shop/config"
	"Shop/database/migrations"
	"Shop/database/models"
	"Shop/loging"
	"Shop/utils"
	"context"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func ShowUserHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	userID, _ := r.Context().Value("userID").(uuid.UUID)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var users []models.User
	cacheKey := "users:all"

	select {
	case <-ctx.Done():
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusRequestTimeout, nil, startTime, "Запрос отменен клиентом")
		return
	default:
	}

	fromCache, err := utils.GetOrSetCache(ctx, config.Rdb, migrations.DB, cacheKey, migrations.DB, &users, 5*time.Minute)
	if err != nil {
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка при поиске пользователя.")
		http.Error(w, "Error fetching couriers", http.StatusInternalServerError)
		return
	}

	if len(users) == 0 {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, nil, startTime, "Пользователи не найдены")
		http.Error(w, "Пользователи не найдены", http.StatusNotFound)
		return
	}
	data := "postgreSQL"
	if fromCache {
		data = "redis"
	}
	utils.JSONFormat(w, r, users)
	loging.LogRequest(logrus.InfoLevel, userID, r, http.StatusOK, nil, startTime, "Список пользователей показан успешно с помощью "+data)
}

func PutMoneyHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	params := mux.Vars(r)
	nickTaker := params["username"]
	userID, _ := r.Context().Value("userID").(uuid.UUID)
	var userTaker, userSender models.User
	tx := migrations.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var walletSender, walletTaker models.Wallet
	if err := tx.Where("user_id = ?", userID).First(&walletSender).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, nil, startTime, "Кошелек работника кто хочет перевести деньги не найден.")
		http.Error(w, "Кошелек работника кто хочет перевести деньги не найден.", http.StatusNotFound)
		return
	}
	if err := tx.Where("user_id = ?", userID).First(&userSender).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, nil, startTime, "Работник кто хочет перевести деньги не найден.")
		http.Error(w, "Работник кто хочет перевести деньги не найден.", http.StatusNotFound)
		return
	}

	if err := tx.Where("username = ?", nickTaker).First(&userTaker).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, nil, startTime, "Работник с никнеймом "+nickTaker+" не найден.")
		http.Error(w, "Работник с никнеймом "+nickTaker+" не найден.", http.StatusNotFound)
		return
	}

	if err := tx.Where("user_id = ?", userTaker.ID).First(&walletTaker).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, nil, startTime, "Кошелек работника которому хотят перевести деньги не найден.")
		http.Error(w, "Кошелек работника которому хотят перевести деньги не найден.", http.StatusNotFound)
		return
	}

	tx.Commit()

}
