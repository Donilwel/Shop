package models_test

import (
	"Shop/database/migrations"
	"Shop/database/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateTransaction(t *testing.T) {
	SetupTestDB()

	transaction := models.Transaction{
		FromUser: uuid.New(),
		ToUser:   uuid.New(),
		Amount:   100,
	}

	result := migrations.DB.Create(&transaction)

	assert.NoError(t, result.Error)
	assert.NotEqual(t, uuid.Nil, transaction.ID)
	assert.Equal(t, uint(100), transaction.Amount)
}

func TestUpdateTransaction(t *testing.T) {
	SetupTestDB()

	transaction := models.Transaction{
		FromUser: uuid.New(),
		ToUser:   uuid.New(),
		Amount:   100,
	}

	migrations.DB.Create(&transaction)

	transaction.Amount = 150
	result := migrations.DB.Save(&transaction)

	assert.NoError(t, result.Error)
	assert.Equal(t, uint(150), transaction.Amount)
}

func TestDeleteTransaction(t *testing.T) {
	SetupTestDB()

	transaction := models.Transaction{
		FromUser: uuid.New(),
		ToUser:   uuid.New(),
		Amount:   100,
	}

	migrations.DB.Create(&transaction)

	result := migrations.DB.Delete(&transaction)

	assert.NoError(t, result.Error)

	var deletedTransaction models.Transaction
	result = migrations.DB.First(&deletedTransaction, transaction.ID)

	assert.Error(t, result.Error)
}

func TestFindTransactionByFromUser(t *testing.T) {
	SetupTestDB()

	fromUser := uuid.New()
	toUser := uuid.New()

	transaction := models.Transaction{
		FromUser: fromUser,
		ToUser:   toUser,
		Amount:   100,
	}

	migrations.DB.Create(&transaction)

	var foundTransaction models.Transaction
	result := migrations.DB.Where("from_user = ?", fromUser).First(&foundTransaction)

	assert.NoError(t, result.Error)
	assert.Equal(t, fromUser, foundTransaction.FromUser)
}

func TestTransactionUUIDGeneration(t *testing.T) {
	SetupTestDB()

	transaction := models.Transaction{
		FromUser: uuid.New(),
		ToUser:   uuid.New(),
		Amount:   100,
	}

	migrations.DB.Create(&transaction)

	assert.NotEqual(t, uuid.Nil, transaction.ID)
}
