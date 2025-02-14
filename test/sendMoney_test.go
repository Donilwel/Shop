package test

//import (
//	"Shop/database/migrations"
//	"Shop/database/models"
//	"fmt"
//	"gorm.io/driver/postgres"
//	"gorm.io/gorm"
//	"os"
//	"testing"
//)
//
//var testDB *gorm.DB
//
//func TestMain(m *testing.M) {
//	var err error
//
//	dsn := fmt.Sprint(
//		"host=postgres user=postgres password=123456 dbname=testdb port=5432 sslmode=disable",
//	)
//
//	testDB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
//	if err != nil {
//		panic("Не удалось подключиться к тестовой БД: " + err.Error())
//	}
//
//	migrations.DB.AutoMigrate(
//		&models.User{},
//		&models.Wallet{},
//		&models.Transaction{},
//	)
//
//	exitCode := m.Run()
//	os.Exit(exitCode)
//}
//
//// Очистка базы перед тестами
//func clearDB() {
//	testDB.Exec("TRUNCATE TABLE transactions, wallets, users RESTART IDENTITY CASCADE;")
//}
