package handlers

import (
	"Shop/config"
	"Shop/database/migrations"
	"Shop/database/models"
	"Shop/loging"
	"Shop/utils"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
	"net/http"
	"time"
)

type Employee struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// ShowEmployeesHandler возвращает список сотрудников.
//
// @Summary Получение списка сотрудников
// @Description Возвращает список сотрудников с их ID, именем пользователя и email из базы данных или кэша Redis.
// @Tags Employee
// @Accept  json
// @Produce  json
// @Success 200 {array} models.User "Список сотрудников"
// @Failure 404 {string} string "Сотрудники не найдены"
// @Failure 408 {string} string "Запрос отменен клиентом"
// @Failure 500 {string} string "Ошибка при поиске сотрудников"
// @Router /api/users [get]
func ShowEmployeesHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	userID, _ := r.Context().Value(utils.UserIDKey).(uuid.UUID)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var users []Employee

	cacheKey := "users:employees"

	select {
	case <-ctx.Done():
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusRequestTimeout, nil, startTime, "Запрос отменен клиентом")
		http.Error(w, "Запрос отменен клиентом", http.StatusRequestTimeout)
		return
	default:
	}

	fromCache, err := utils.GetOrSetCache(ctx, config.Rdb, migrations.DB, cacheKey,
		migrations.DB.Model(&models.User{}).
			Select("id, username, email").
			Where("role = ?", models.EMPLOYEE_ROLE), &users, 5*time.Minute)
	if err != nil {
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка при поиске сотрудников.")
		http.Error(w, "Ошибка при поиске сотрудников", http.StatusInternalServerError)
		return
	}

	if len(users) == 0 {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, nil, startTime, "Сотрудники не найдены")
		http.Error(w, "Сотрудники не найдены", http.StatusNotFound)
		return
	}

	data := "postgreSQL"
	if fromCache {
		data = "redis"
	}

	utils.JSONFormat(w, r, users)
	loging.LogRequest(logrus.InfoLevel, userID, r, http.StatusOK, nil, startTime, "Список сотрудников показан успешно с помощью "+data)
}

type InfoMain struct {
	Coins     uint `json:"coins"`
	Inventory []struct {
		Type     string `json:"type"`
		Quantity int    `json:"quantity"`
	} `json:"inventory"`
	CoinHistory struct {
		Received []struct {
			FromUser string `json:"fromUser"`
			Amount   uint   `json:"amount"`
		} `json:"received"`
		Sent []struct {
			ToUser string `json:"toUser"`
			Amount uint   `json:"amount"`
		} `json:"sent"`
	} `json:"coinHistory"`
}

// InformationHandler информация о пользователе
//
// @Summary Получение информации о кошельке, инвентаре и транзакциях пользователя
// @Description Возвращает информацию о кошельке, инвентаре и истории транзакций для конкретного пользователя.
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {token}"
// @Success 200 {object} InfoMain "Информация о кошельке и транзакциях"
// @Failure 400 {object} string "Некорректный запрос"
// @Failure 404 {object} string "Не найден пользователь или его данные"
// @Failure 500 {object} string "Ошибка на сервере при получении данных"
// @Router /api/info [get]
// @Security BearerAuth
func InformationHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	ctx := r.Context()
	userID, _ := ctx.Value(utils.UserIDKey).(uuid.UUID)
	cacheTTL := 5 * time.Minute

	walletCacheKey := fmt.Sprintf("wallet:%s", userID)
	inventoryCacheKey := fmt.Sprintf("inventory:%s", userID)
	receivedCacheKey := fmt.Sprintf("received:%s", userID)
	sentCacheKey := fmt.Sprintf("sent:%s", userID)

	var walletSlice []models.Wallet
	fromCacheWallet, err := utils.GetOrSetCache(ctx, config.Rdb, migrations.DB, walletCacheKey,
		migrations.DB.Where("user_id = ?", userID), &walletSlice, cacheTTL)
	if err != nil {
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка при получении кошелька")
		http.Error(w, "Failed to retrieve wallet", http.StatusInternalServerError)
		return
	}
	var wallet models.Wallet
	if len(walletSlice) > 0 {
		wallet = walletSlice[0]
	}
	loging.LogRequest(logrus.InfoLevel, userID, r, http.StatusOK, nil, startTime, "Кошелек загружен из "+GetSource(fromCacheWallet))

	var inventory []struct {
		Type     string `json:"type"`
		Quantity int    `json:"quantity"`
	}
	fromCacheInventory, err := utils.GetOrSetCache(ctx, config.Rdb, migrations.DB, inventoryCacheKey,
		migrations.DB.Table("purchases").
			Select("merches.name as type, COUNT(purchases.id) as quantity").
			Joins("JOIN merches ON purchases.merch_id = merches.id").
			Where("purchases.user_id = ?", userID).
			Group("merches.id, merches.name"), &inventory, cacheTTL)
	if err != nil {
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка при получении инвентаря")
		http.Error(w, "Failed to retrieve inventory", http.StatusInternalServerError)
		return
	}
	loging.LogRequest(logrus.InfoLevel, userID, r, http.StatusOK, nil, startTime, "Инвентарь загружен из "+GetSource(fromCacheInventory))

	var received []struct {
		FromUser string `json:"fromUser"`
		Amount   uint   `json:"amount"`
	}
	fromCacheReceived, err := utils.GetOrSetCache(ctx, config.Rdb, migrations.DB, receivedCacheKey,
		migrations.DB.Table("transactions").
			Select("users.username as from_user, transactions.amount").
			Joins("JOIN users ON transactions.from_user = users.id").
			Where("transactions.to_user = ?", userID), &received, cacheTTL)
	if err != nil {
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка при получении полученных транзакций")
		http.Error(w, "Failed to retrieve received transactions", http.StatusInternalServerError)
		return
	}
	loging.LogRequest(logrus.InfoLevel, userID, r, http.StatusOK, nil, startTime, "Отправленные транзакции загружены из "+GetSource(fromCacheReceived))

	var sent []struct {
		ToUser string `json:"toUser"`
		Amount uint   `json:"amount"`
	}
	fromCacheSent, err := utils.GetOrSetCache(ctx, config.Rdb, migrations.DB, sentCacheKey,
		migrations.DB.Table("transactions").
			Select("users.username as to_user, transactions.amount").
			Joins("JOIN users ON transactions.to_user = users.id").
			Where("transactions.from_user = ?", userID), &sent, cacheTTL)
	if err != nil {
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка при получении отправленных транзакций")
		http.Error(w, "Failed to retrieve sent transactions", http.StatusInternalServerError)
		return
	}

	loging.LogRequest(logrus.InfoLevel, userID, r, http.StatusOK, nil, startTime, "Отправленные транзакции загружены из "+GetSource(fromCacheSent))

	response := InfoMain{
		Coins:     wallet.Coin,
		Inventory: inventory,
		CoinHistory: struct {
			Received []struct {
				FromUser string `json:"fromUser"`
				Amount   uint   `json:"amount"`
			} `json:"received"`
			Sent []struct {
				ToUser string `json:"toUser"`
				Amount uint   `json:"amount"`
			} `json:"sent"`
		}{
			Received: received,
			Sent:     sent,
		},
	}

	utils.JSONFormat(w, r, response)
	loging.LogRequest(logrus.InfoLevel, userID, r, http.StatusOK, nil, startTime, "Информация показана успешно")
}

func GetSource(fromCache bool) string {
	if fromCache {
		return "Redis"
	}
	return "PostgreSQL"
}

type TransactionsResponse struct {
	NickTaker string `json:"toUser"`
	Coin      uint   `json:"coin"`
}

// SendCoinHandler Отправка монет
// @Summary Отправка монет от одного пользователя другому
// @Description Позволяет пользователю отправить монеты другому пользователю, указав его имя и количество монет для отправки.
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {token}"
// @Param request body TransactionsResponse true "Тело запроса"
// @Success 200 {object} models.Transaction "Транзакция успешно создана"
// @Failure 400 {object} string "Неверный запрос - некорректный ввод, недостаточно монет или попытка отправки себе"
// @Failure 404 {object} string "Не найдено - пользователь или кошелек не найдены"
// @Failure 500 {object} string "Внутренняя ошибка сервера - проблемы с транзакцией в базе данных"
// @Router /api/sendCoin [post]
// @Security BearerAuth
func SendCoinHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	userID, _ := r.Context().Value(utils.UserIDKey).(uuid.UUID)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var input = TransactionsResponse{}

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
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка обновления баланса отправителя")
		http.Error(w, "Ошибка обновления баланса отправителя.", http.StatusInternalServerError)
		return
	}
	if err := tx.WithContext(ctx).Save(&walletTaker).Error; err != nil {
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка обновления баланса получателя")
		http.Error(w, "Ошибка обновления баланса получателя.", http.StatusInternalServerError)
		return
	}

	transaction := models.Transaction{
		FromUser: userSender.ID,
		ToUser:   userTaker.ID,
		Amount:   input.Coin,
	}
	if err := tx.WithContext(ctx).Create(&transaction).Error; err != nil {
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка создания транзакции")
		http.Error(w, "Ошибка создания транзакции.", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка фиксации транзакции")
		http.Error(w, "Ошибка фиксации транзакции.", http.StatusInternalServerError)
		return
	}

	committed = true
	utils.JSONFormat(w, r, transaction)
}

type InfoAfterBying struct {
	Balance  interface{} `json:"balance"`
	Item     interface{} `json:"item"`
	Nickname interface{} `json:"nickname"`
}

// BuyItemHandler Покупка товара
// @Summary Покупка товара пользователем
// @Description Позволяет пользователю купить товар, указав его имя. Проверяется наличие средств на кошельке и успешность покупки.
// @Tags Employee
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {token}"
// @Param item path string true "Название товара" example("item_name")
// @Success 200 {object} InfoAfterBying "Информация о балансе и купленном товаре"
// @Failure 400 {object} string "Недостаточно средств на кошельке"
// @Failure 404 {object} string "Покупатель или товар не найдены"
// @Failure 500 {object} string "Ошибка сохранения в базе данных"
// @Router /api/buy/{item} [get]
// @Security BearerAuth
func BuyItemHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	userID := r.Context().Value(utils.UserIDKey).(uuid.UUID)

	itemName := mux.Vars(r)["item"]
	var merch models.Merch
	var user models.User

	tx := migrations.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Where("id = ?", userID).First(&user).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, err, startTime, "Покупатель не найден в базе данных")
		http.Error(w, "Покупатель не найден в базе данных", http.StatusNotFound)
		return
	}
	var wallet models.Wallet
	if err := tx.Where("user_id = ?", userID).First(&wallet).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, err, startTime, "Кошелька покупателя не существует в базе данных")
		http.Error(w, "Кошелька покупателя не существует в базе данных", http.StatusNotFound)
		return
	}

	if err := tx.Where("name = ?", itemName).First(&merch).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, err, startTime, "Запрошенная вещь не существует в базе данных")
		http.Error(w, "Запрошенная вещь не существует в базе данных", http.StatusNotFound)
		return
	}

	if wallet.Coin < merch.Price {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusBadRequest, nil, startTime, "Недостаточно средств на кошельке у пользователя.")
		http.Error(w, "Недостаточно средств на кошельке у пользователя.", http.StatusBadRequest)
		return
	}

	wallet.Coin -= merch.Price
	var purchase = models.Purchase{
		UserID:  userID,
		MerchID: merch.ID,
	}

	if err := tx.Save(&purchase).Error; err != nil {
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка сохранения в истории заказа.")
		http.Error(w, "Ошибка сохранения в истории заказа.", http.StatusInternalServerError)
		return
	}
	if err := tx.Save(&wallet).Error; err != nil {
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка сохранения кошелька у пользователя.")
		http.Error(w, "Ошибка сохранения кошелька у пользователя.", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка фиксации транзакции")
		http.Error(w, "Ошибка фиксации транзакции", http.StatusInternalServerError)
		return
	}
	utils.JSONFormat(w, r, InfoAfterBying{
		Balance:  wallet.Coin,
		Item:     itemName,
		Nickname: user.Username,
	})
}
