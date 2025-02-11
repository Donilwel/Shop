package handlers

import (
	"Shop/database/migrations"
	"Shop/database/models"
	"Shop/utils"
	"github.com/google/uuid"
	"net/http"
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

func SendCoinHandler(w http.ResponseWriter, r *http.Request) {}

func BuyItemHandler(w http.ResponseWriter, r *http.Request) {}
