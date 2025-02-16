package models_test

import (
	"Shop/database/migrations"
	"Shop/database/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCreatePurchase(t *testing.T) {
	SetupTestDB()

	purchase := models.Purchase{
		UserID:  uuid.New(),
		MerchID: uuid.New(),
	}

	result := migrations.DB.Create(&purchase)

	assert.NoError(t, result.Error)
	assert.NotEqual(t, uuid.Nil, purchase.ID)
	assert.NotNil(t, purchase.CreatedAt)
}

func TestUpdatePurchase(t *testing.T) {
	SetupTestDB()

	userID := uuid.New()
	merchID := uuid.New()

	purchase := models.Purchase{
		UserID:  userID,
		MerchID: merchID,
	}
	migrations.DB.Create(&purchase)

	newCreatedAt := time.Now().Add(24 * time.Hour)
	purchase.CreatedAt = newCreatedAt
	migrations.DB.Save(&purchase)

	var updatedPurchase models.Purchase
	result := migrations.DB.First(&updatedPurchase, purchase.ID)

	assert.NoError(t, result.Error)
	assert.Equal(t, newCreatedAt.Format(time.RFC3339), updatedPurchase.CreatedAt.Format(time.RFC3339))
}

func TestFindPurchaseByUserID(t *testing.T) {
	SetupTestDB()

	userID := uuid.New()
	merchID := uuid.New()

	purchase := models.Purchase{
		UserID:  userID,
		MerchID: merchID,
	}
	migrations.DB.Create(&purchase)

	var foundPurchase models.Purchase
	result := migrations.DB.Where("user_id = ?", userID).First(&foundPurchase)

	assert.NoError(t, result.Error)
	assert.Equal(t, userID, foundPurchase.UserID)
}

func TestPurchaseUUIDGeneration(t *testing.T) {
	SetupTestDB()

	userID := uuid.New()
	merchID := uuid.New()

	purchase := models.Purchase{
		UserID:  userID,
		MerchID: merchID,
	}

	migrations.DB.Create(&purchase)

	assert.NotEqual(t, uuid.Nil, purchase.ID)
}
