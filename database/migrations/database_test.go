package migrations_test

import (
	"Shop/database/models"
	"Shop/loging"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		loging.Log.Fatal("Ошибка подключения к тестовой БД")
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Merch{},
		&models.RevokedToken{},
		&models.Transaction{},
		&models.Purchase{},
		&models.Wallet{},
	); err != nil {
		loging.Log.Fatal("Ошибка миграции тестовой БД")
	}

	return db
}

//func TestInitDB_Success(t *testing.T) {
//	os.Setenv("POSTGRES_HOST", "localhost")
//	os.Setenv("POSTGRES_USERNAME", "test_user")
//	os.Setenv("POSTGRES_PASSWORD", "test_pass")
//	os.Setenv("POSTGRES_DATABASE", "test_db")
//	os.Setenv("POSTGRES_PORT", "5432")
//
//	testDB := setupTestDB()
//	migrations.DB = testDB
//
//	assert.NotNil(t, migrations.DB, "База данных должна быть инициализирована")
//}
//
//func TestAutoMigrate_Success(t *testing.T) {
//	testDB := setupTestDB()
//	migrations.DB = testDB
//
//	err := testDB.AutoMigrate(&models.User{}, &models.Merch{}, &models.RevokedToken{}, &models.Transaction{}, &models.Purchase{}, &models.Wallet{})
//	assert.NoError(t, err, "Миграция должна пройти без ошибок")
//}
