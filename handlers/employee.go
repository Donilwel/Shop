package handlers

import (
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

func InformationHandler(w http.ResponseWriter, r *http.Request) {
	userID, _ := r.Context().Value("userID").(uuid.UUID)

	var wallet models.Wallet
	if err := migrations.DB.Where("user_id = ?", userID).First(&wallet).Error; err != nil {

	}
	var purchases []models.Purchase
	if err := migrations.DB.Where("user_id = ?", userID).Find(&purchases).Error; err != nil {

	}
	var transactionsFrom []models.Transaction
	var transactionsTo []models.Transaction
	if err := migrations.DB.Where("from_user = ?", userID).Find(&transactionsFrom).Error; err != nil {

	}
	if err := migrations.DB.Where("to_user = ?", userID).Find(&transactionsTo).Error; err != nil {

	}
	utils.JSONFormat(w, r, map[string]interface{}{
		"coins":             wallet.Coin,
		"purchases":         purchases,
		"transactions":      transactionsTo,
		"transactions_from": transactionsFrom,
	})
}

func SendCoinHandler(w http.ResponseWriter, r *http.Request) {
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

	if input.Coin == 0 {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusBadRequest, nil, startTime, "Количество монет должно быть больше 0")
		http.Error(w, "Количество монет должно быть больше 0", http.StatusBadRequest)
		return
	}

	tx := migrations.DB.Begin()
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
		}
	}()

	var walletSender, walletTaker models.Wallet
	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userID).First(&walletSender).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, err, startTime, "Кошелек отправителя не найден")
		http.Error(w, "Кошелек отправителя не найден.", http.StatusNotFound)
		return
	}

	if input.Coin > walletSender.Coin {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusBadRequest, nil, startTime, "Недостаточно монет на балансе")
		http.Error(w, "Недостаточно монет на балансе.", http.StatusBadRequest)
		return
	}

	var userSender, userTaker models.User
	if err := tx.WithContext(ctx).Where("id = ?", userID).First(&userSender).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, err, startTime, "Отправитель не найден")
		http.Error(w, "Отправитель не найден.", http.StatusNotFound)
		return
	}
	if err := tx.WithContext(ctx).Where("username = ?", input.NickTaker).First(&userTaker).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, err, startTime, "Получатель не найден")
		http.Error(w, "Получатель не найден.", http.StatusNotFound)
		return
	}

	if userSender.Username == userTaker.Username {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusBadRequest, nil, startTime, "Самому себе нельзя перевести деньги")
		http.Error(w, "Самому себе нельзя перевести деньги.", http.StatusBadRequest)
		return
	}

	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userTaker.ID).First(&walletTaker).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, err, startTime, "Кошелек получателя не найден")
		http.Error(w, "Кошелек получателя не найден.", http.StatusNotFound)
		return
	}

	walletSender.Coin -= input.Coin
	walletTaker.Coin += input.Coin

	if err := tx.WithContext(ctx).Save(&walletSender).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка обновления баланса отправителя")
		http.Error(w, "Ошибка обновления баланса отправителя.", http.StatusInternalServerError)
		return
	}
	if err := tx.WithContext(ctx).Save(&walletTaker).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка обновления баланса получателя")
		http.Error(w, "Ошибка обновления баланса получателя.", http.StatusInternalServerError)
		return
	}

	transaction := models.Transaction{
		FromUser: userSender.ID,
		ToUser:   userTaker.ID,
		Amount:   input.Coin,
	}
	if err := tx.WithContext(ctx).Create(&transaction).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка создания транзакции")
		http.Error(w, "Ошибка создания транзакции.", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка фиксации транзакции")
		http.Error(w, "Ошибка фиксации транзакции.", http.StatusInternalServerError)
		return
	}

	committed = true
	utils.JSONFormat(w, r, transaction)
}

func BuyItemHandler(w http.ResponseWriter, r *http.Request) {}
