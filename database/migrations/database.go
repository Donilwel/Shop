package migrations

import (
	"Shop/database/models"
	"Shop/loging"
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
)

var DB *gorm.DB

func InitDB() {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_USERNAME"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DATABASE"),
		os.Getenv("POSTGRES_PORT"),
	)
	fmt.Println("DSN:", dsn)

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		loging.Log.WithError(err).Fatal("Ошибка подключения к PostgreSQL базе данных")
	}

	if err := DB.AutoMigrate(
		&models.User{},
		&models.Merch{},
		&models.RevokedToken{},
		&models.Transaction{},
		&models.Purchase{},
		&models.Wallet{},
	); err != nil {
		loging.Log.WithError(err).Fatal("Ошибка миграции базы данных.")
		return
	}
	loging.Log.Info("database migrated successfully")
}
