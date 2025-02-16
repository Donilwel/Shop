package models_test

import (
	"Shop/database/migrations"
	"Shop/database/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateUser(t *testing.T) {
	SetupTestDB()

	user := models.User{
		Username: "testuser",
		Email:    "testuser@example.com",
		Password: "password123",
		Role:     "ADMIN_ROLE",
	}

	result := migrations.DB.Create(&user)

	assert.NoError(t, result.Error)
	assert.NotEqual(t, uuid.Nil, user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "testuser@example.com", user.Email)
	assert.Equal(t, "ADMIN_ROLE", user.Role)
}

func TestCreateUserWithDuplicateUsername(t *testing.T) {
	SetupTestDB()

	user1 := models.User{
		Username: "testuser",
		Email:    "testuser@example.com",
		Password: "password123",
	}
	migrations.DB.Create(&user1)

	user2 := models.User{
		Username: "testuser",
		Email:    "testuser2@example.com",
		Password: "password123",
	}

	result := migrations.DB.Create(&user2)

	assert.Error(t, result.Error)
}

func TestUpdateUserRole(t *testing.T) {
	SetupTestDB()

	user := models.User{
		Username: "testuser",
		Email:    "testuser@example.com",
		Password: "password123",
	}

	migrations.DB.Create(&user)

	user.Role = "EMPLOYEE_ROLE"
	result := migrations.DB.Save(&user)

	assert.NoError(t, result.Error)
	assert.Equal(t, "EMPLOYEE_ROLE", user.Role)
}

func TestDeleteUser(t *testing.T) {
	SetupTestDB()

	user := models.User{
		Username: "testuser",
		Email:    "testuser@example.com",
		Password: "password123",
	}

	migrations.DB.Create(&user)

	result := migrations.DB.Delete(&user)

	assert.NoError(t, result.Error)

	var deletedUser models.User
	result = migrations.DB.First(&deletedUser, user.ID)

	assert.Error(t, result.Error)
}

func TestFindUserByEmail(t *testing.T) {
	SetupTestDB()

	user := models.User{
		Username: "testuser",
		Email:    "testuser@example.com",
		Password: "password123",
	}

	migrations.DB.Create(&user)

	var foundUser models.User
	result := migrations.DB.Where("email = ?", "testuser@example.com").First(&foundUser)

	assert.NoError(t, result.Error)
	assert.Equal(t, "testuser@example.com", foundUser.Email)
}

func TestUserUUIDGeneration(t *testing.T) {
	SetupTestDB()

	user := models.User{
		Username: "testuser",
		Email:    "testuser@example.com",
		Password: "password123",
	}

	migrations.DB.Create(&user)

	assert.NotEqual(t, uuid.Nil, user.ID)
}
