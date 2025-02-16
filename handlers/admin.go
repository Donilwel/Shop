package handlers

import (
	"Shop/database/migrations"
	"Shop/database/models"
	"Shop/loging"
	"Shop/utils"
	"context"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm/clause"
	"net/http"
	"time"
)

type SendMoney struct {
	NickTaker string `json:"toUser"`
	Coin      uint   `json:"coin"`
}

// PutMoneyHandler Перевод монет работнику
//
// @Summary Перевод монет работнику
// @Description Позволяет перевести монеты работнику по его никнейму, проверяя корректность данных и существование получателя.
// @Tags Admin
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {token}"
// @Param request body SendMoney true "Тело запроса"
// @Success 200 {object} string "Перевод монет успешен"
// @Failure 400 {object} string "Некорректное тело запроса или неверное количество монет"
// @Failure 404 {object} string "Не найден работник или кошелек получателя"
// @Failure 500 {object} string "Ошибка обновления баланса получателя или фиксации транзакции"
// @Router /api/admin/users [post]
// @Security BearerAuth
func PutMoneyHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	userID, _ := r.Context().Value(utils.UserIDKey).(uuid.UUID)

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var input SendMoney

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusBadRequest, err, startTime, "Некорректное тело запроса")
		http.Error(w, "Некорректное тело запроса", http.StatusBadRequest)
		return
	}

	if input.Coin == 0 || input.Coin > 1000 {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusBadRequest, nil, startTime, "Количество монет должно быть в диапазоне от 1 до 1000 включительно")
		http.Error(w, "Количество монет должно быть в диапазоне от 1 до 1000 включительно", http.StatusBadRequest)
		return
	}

	var userTaker models.User
	tx := migrations.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var walletTaker models.Wallet

	if err := tx.Where("username = ?", input.NickTaker).First(&userTaker).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, nil, startTime, "Работник с никнеймом "+input.NickTaker+" не найден.")
		http.Error(w, "Работник с никнеймом "+input.NickTaker+" не найден.", http.StatusNotFound)
		return
	}

	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", userTaker.ID).First(&walletTaker).Error; err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusNotFound, err, startTime, "Кошелек получателя не найден")
		http.Error(w, "Кошелек получателя не найден.", http.StatusNotFound)
		return
	}
	walletTaker.Coin += input.Coin

	if err := tx.WithContext(ctx).Save(&walletTaker).Error; err != nil {
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка обновления баланса получателя")
		http.Error(w, "Ошибка обновления баланса получателя.", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка фиксации транзакции")
		http.Error(w, "Ошибка фиксации транзакции", http.StatusInternalServerError)
		return
	}
	http.Error(w, "Перевод монет успешен", http.StatusOK)
}

type MerchInfo struct {
	Type  string `json:"type"`
	Price uint   `json:"price"`
}

// AddOrChangeMerchHandler добавить или изменить цену мерча
//
// @Summary Добавление или изменение цены мерча
// @Description Позволяет добавить новый мерч или изменить цену существующего мерча. Проверяет корректность данных и наличие мерча.
// @Tags Admin
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {token}"
// @Param request body MerchInfo true "Тело запроса"
// @Success 200 {object} string "Мерч успешно добавлен или цена обновлена"
// @Failure 400 {object} string "Некорректное тело запроса, неверный тип или цена мерча"
// @Failure 404 {object} string "Мерч с таким именем уже существует"
// @Failure 500 {object} string "Ошибка добавления нового мерча или обновления цены"
// @Router /api/admin/merch/new [post]
// @Security BearerAuth
func AddOrChangeMerchHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	userID, ok := r.Context().Value(utils.UserIDKey).(uuid.UUID)
	if !ok {
		http.Error(w, "Не удалось получить userID", http.StatusUnauthorized)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var input MerchInfo

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusBadRequest, err, startTime, "Некорректное тело запроса")
		http.Error(w, "Некорректное тело запроса", http.StatusBadRequest)
		return
	}

	if input.Type == "" {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusBadRequest, nil, startTime, "Тип мерча не должен быть пустым")
		http.Error(w, "Тип мерча не должен быть пустым", http.StatusBadRequest)
		return
	}

	if input.Price == 0 || input.Price > 1000 {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusBadRequest, nil, startTime, "Цена мерча должна быть в диапазоне от 1 до 1000 включительно")
		http.Error(w, "Цена мерча должна быть в диапазоне от 1 до 1000 включительно", http.StatusBadRequest)
		return
	}

	tx := migrations.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var merchExist models.Merch
	if err := tx.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("name = ?", input.Type).First(&merchExist).Error; err != nil {
		merch := models.Merch{
			Name:  input.Type,
			Price: input.Price,
		}

		if err := tx.WithContext(ctx).Create(&merch).Error; err != nil {
			tx.Rollback()
			loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка добавления нового мерча")
			http.Error(w, "Ошибка добавления нового мерча", http.StatusInternalServerError)
			return
		}

		loging.LogRequest(logrus.InfoLevel, userID, r, http.StatusOK, nil, startTime, "Был создан новый мерч: "+input.Type)
		http.Error(w, "Был создан новый мерч: "+input.Type, http.StatusOK)
	} else {
		if merchExist.Price == input.Price {
			loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusBadRequest, nil, startTime, "цена мерча совпадает с заданной")
			http.Error(w, "цена мерча совпадает с заданной", http.StatusBadRequest)
			return
		}
		merchExist.Price = input.Price

		if err := tx.WithContext(ctx).Save(&merchExist).Error; err != nil {
			tx.Rollback()
			loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка обновления цены мерча")
			http.Error(w, "Ошибка обновления цены мерча", http.StatusInternalServerError)
			return
		}

		loging.LogRequest(logrus.InfoLevel, userID, r, http.StatusOK, nil, startTime, "Цена мерча "+input.Type+" была обновлена")
		http.Error(w, "Цена мерча "+input.Type+" была обновлена", http.StatusOK)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Ошибка фиксации транзакции")
		http.Error(w, "Ошибка фиксации транзакции", http.StatusInternalServerError)
		return
	}
}
