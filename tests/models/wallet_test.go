package models_test

import (
	"Shop/database/migrations"
	"Shop/database/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateWallet(t *testing.T) {
	SetupTestDB()

	userID := uuid.New()

	wallet := models.Wallet{
		UserID: userID,
		Coin:   1000,
	}

	result := migrations.DB.Create(&wallet)

	assert.NoError(t, result.Error)
	assert.NotEqual(t, uuid.Nil, wallet.ID)
	assert.Equal(t, userID, wallet.UserID)
	assert.Equal(t, uint(1000), wallet.Coin)
}

func TestUpdateWallet(t *testing.T) {
	SetupTestDB()

	userID := uuid.New()

	wallet := models.Wallet{
		UserID: userID,
		Coin:   1000,
	}

	migrations.DB.Create(&wallet)

	wallet.Coin = 1500
	result := migrations.DB.Save(&wallet)

	assert.NoError(t, result.Error)
	assert.Equal(t, uint(1500), wallet.Coin)
}

func TestDeleteWallet(t *testing.T) {
	SetupTestDB()

	userID := uuid.New()

	wallet := models.Wallet{
		UserID: userID,
		Coin:   1000,
	}

	migrations.DB.Create(&wallet)

	result := migrations.DB.Delete(&wallet)

	assert.NoError(t, result.Error)

	var deletedWallet models.Wallet
	result = migrations.DB.First(&deletedWallet, wallet.ID)

	assert.Error(t, result.Error)
}

func TestFindWalletByUserID(t *testing.T) {
	SetupTestDB()

	userID := uuid.New()

	wallet := models.Wallet{
		UserID: userID,
		Coin:   1000,
	}

	migrations.DB.Create(&wallet)

	var foundWallet models.Wallet
	result := migrations.DB.Where("user_id = ?", userID).First(&foundWallet)

	assert.NoError(t, result.Error)
	assert.Equal(t, userID, foundWallet.UserID)
}

func TestWalletUUIDGeneration(t *testing.T) {
	SetupTestDB()

	userID := uuid.New()

	wallet := models.Wallet{
		UserID: userID,
		Coin:   1000,
	}

	migrations.DB.Create(&wallet)

	assert.NotEqual(t, uuid.Nil, wallet.ID)
}
