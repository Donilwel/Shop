package handlers

import (
	"Shop/database/migrations"
	"Shop/database/models"
	"Shop/loging"
	"Shop/utils"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
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

	var input struct {
		NickTaker string `json:"toUser"`
		Coin      uint   `json:"coin"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusBadRequest, err, startTime, "Некорректное тело запроса")
		http.Error(w, "Некорректное тело запроса", http.StatusBadRequest)
		return
	}

	var userTaker, userSender models.User
	tx := migrations.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var walletSender, walletTaker models.Wallet
	if err := tx.Where("user_id = ?", userID).First(&walletSender).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, err, startTime, "Кошелек работника кто хочет перевести деньги не найден.")
		http.Error(w, "Кошелек работника кто хочет перевести деньги не найден.", http.StatusNotFound)
		return
	}

	if input.Coin <= 0 {
		loging.LogRequest(logrus.WarnLevel, uuid.Nil, r, http.StatusBadRequest, nil, startTime, "Количество монет должно быть больше 0")
		http.Error(w, "Количество монет должно быть больше 0", http.StatusBadRequest)
		return
	}
	if input.Coin > walletSender.Coin {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusBadRequest, nil, startTime, "Количество денег у работника кто хочет перевести деньги недостаточно.")
		http.Error(w, "Количество денег у работника кто хочет перевести деньги недостаточно.", http.StatusBadRequest)
		return
	}

	if err := tx.Where("id = ?", userID).First(&userSender).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, err, startTime, "Работник кто хочет перевести деньги не найден.")
		http.Error(w, "Работник кто хочет перевести деньги не найден.", http.StatusNotFound)
		return
	}

	if err := tx.Where("username = ?", input.NickTaker).First(&userTaker).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, err, startTime, "Работник с никнеймом "+input.NickTaker+" не найден.")
		http.Error(w, "Работник с никнеймом "+input.NickTaker+" не найден.", http.StatusNotFound)
		return
	}
	if userSender.Username == userTaker.Username {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusBadRequest, nil, startTime, "Самому себе нельзя перевести деньги.")
		http.Error(w, "Самому себе нельзя перевести деньги.", http.StatusBadRequest)
		return
	}

	if err := tx.Where("user_id = ?", userTaker.ID).First(&walletTaker).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, err, startTime, "Кошелек работника которому хотят перевести деньги не найден.")
		http.Error(w, "Кошелек работника которому хотят перевести деньги не найден.", http.StatusNotFound)
		return
	}

	walletSender.Coin -= input.Coin
	walletTaker.Coin += input.Coin

	var transaction = models.Transaction{
		FromUser: userSender.ID,
		ToUser:   userTaker.ID,
		Amount:   input.Coin,
	}
	if err := tx.Create(&transaction).Error; err != nil {
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка при создании транзакции перевода денег.")
		http.Error(w, "Ошибка при создании транзакции перевода денег.", http.StatusInternalServerError)
		return
	}

	if err := tx.Save(&walletSender).Error; err != nil {
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка при сохранении баланса кошелька у отправителя.")
		http.Error(w, "Ошибка при сохранении баланса кошелька у отправителя.", http.StatusInternalServerError)
		return
	}

	if err := tx.Save(&walletTaker).Error; err != nil {
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка при сохранении баланса кошелька у получателя.")
		http.Error(w, "Ошибка при сохранении баланса кошелька у получателя.", http.StatusInternalServerError)
		return
	}

	tx.Commit()
	utils.JSONFormat(w, r, transaction)
}

func BuyItemHandler(w http.ResponseWriter, r *http.Request) {}
