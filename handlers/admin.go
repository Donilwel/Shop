package handlers

import (
	"Shop/config"
	"Shop/database/migrations"
	"Shop/database/models"
	"Shop/loging"
	"Shop/utils"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
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
	userID, _ := r.Context().Value("userID").(uuid.UUID)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var input struct {
		NickTaker string `json:"toUser"`
		Coin      uint   `json:"coin"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusBadRequest, err, startTime, "Некорректное тело запроса")
		http.Error(w, "Некорректное тело запроса", http.StatusBadRequest)
		return
	}

	if input.Coin == 0 || input.Coin > 1000 {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusBadRequest, nil, startTime, "Количество монет должно быть в диапазоне от 1 до 1000 включительно")
		http.Error(w, "Количество монет должно быть в диапазоне от 1 до 1000 включительно", http.StatusBadRequest)
		return
	}

	var userTaker models.User
	tx := migrations.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var walletTaker models.Wallet

	if err := tx.Where("username = ?", input.NickTaker).First(&userTaker).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, nil, startTime, "Работник с никнеймом "+input.NickTaker+" не найден.")
		http.Error(w, "Работник с никнеймом "+input.NickTaker+" не найден.", http.StatusNotFound)
		return
	}

	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userTaker.ID).First(&walletTaker).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, err, startTime, "Кошелек получателя не найден")
		http.Error(w, "Кошелек получателя не найден.", http.StatusNotFound)
		return
	}
	walletTaker.Coin += input.Coin

	if err := tx.WithContext(ctx).Save(&walletTaker).Error; err != nil {
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка обновления баланса получателя")
		http.Error(w, "Ошибка обновления баланса получателя.", http.StatusInternalServerError)
		return
	}

	tx.Commit()
}

func AddOrChangeMerchHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	userID, _ := r.Context().Value("userID").(uuid.UUID)
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var input struct {
		Type  string `json:"type"`
		Price uint   `json:"price"`
	}
	if input.Type == "" {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusBadRequest, nil, startTime, "Тип мерча не должен быть пустым")
		http.Error(w, "Тип мерча не должен быть пустым", http.StatusBadRequest)
		return
	}
	if input.Price == 0 || input.Price > 1000 {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusBadRequest, nil, startTime, "Цена мерча должна быть в диапазоне от 1 до 1000 включительно")
		http.Error(w, "Цена мерча должна быть в диапазоне от 1 до 1000 включительно", http.StatusBadRequest)
		return
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusBadRequest, err, startTime, "Некорректное тело запроса")
		http.Error(w, "Некорректное тело запроса", http.StatusBadRequest)
		return
	}

	tx := migrations.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	// тут надо будет добавить if мерч с типом существует то тогда просто апдейтнуть цену - иначе - то что ниже
	var merch = models.Merch{
		Name:  input.Type,
		Price: input.Price,
	}

	if err := tx.WithContext(ctx).Create(&merch).Error; err != nil {
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка добавления нового мерча.")
		http.Error(w, "Ошибка добавления нового мерча.", http.StatusInternalServerError)
		return
	}
}
