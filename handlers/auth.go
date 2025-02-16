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
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"
)

const (
	ADMIN_EMAIL    string = "admin@admin"
	ADMIN_PASSWORD string = "$2a$10$W6GQy6eZ6GjyrQzKZwjmn.wQ9WT4QZ7lpq3Kv6/G1Fdq16VHPcITm"
)

// AuthRequest представляет тело запроса для авторизации.
//
// @Description Структура для входа пользователя
type AuthRequest struct {
	Email    string `json:"email" example:"user@example.com"`
	Password string `json:"password" example:"securepassword"`
}

// AuthHandler обрабатывает аутентификацию пользователя.
//
// @Summary Авторизация пользователя
// @Description Авторизует пользователя по email и паролю, создавая учетную запись автоматически, если пользователя нет в базе.
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param request body AuthRequest true "Тело запроса"
// @Success 200 {object} map[string]string "Возвращает JWT-токен"
// @Success 201 {string} string "Пользователь создан автоматически"
// @Failure 400 {string} string "Некорректное тело запроса"
// @Failure 401 {string} string "Неверный пароль"
// @Failure 408 {string} string "Запрос отменен клиентом"
// @Failure 500 {string} string "Внутренняя ошибка сервера"
// @Router /api/auth [post]
func AuthHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var input AuthRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		loging.LogRequest(logrus.WarnLevel, uuid.Nil, r, http.StatusBadRequest, err, startTime, "Некорректное тело запроса")
		http.Error(w, "Некорректное тело запроса", http.StatusBadRequest)
		return
	}
	tx := migrations.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	select {
	case <-ctx.Done():
		loging.LogRequest(logrus.WarnLevel, uuid.Nil, r, http.StatusRequestTimeout, nil, startTime, "Запрос отменен клиентом")
		http.Error(w, "Запрос отменен", http.StatusRequestTimeout)
		return
	default:
	}

	var user models.User
	if err := tx.WithContext(ctx).Where("email = ?", input.Email).First(&user).Error; err != nil {
		if input.Email == "" || input.Password == "" {
			loging.LogRequest(logrus.WarnLevel, uuid.Nil, r, http.StatusBadRequest, nil, startTime, "Email пользователя обязателен при первой авторизации")
			http.Error(w, "Email пользователя обязательно при первой авторизации", http.StatusBadRequest)
			return
		}

		hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if hashErr != nil {
			loging.LogRequest(logrus.ErrorLevel, uuid.Nil, r, http.StatusInternalServerError, hashErr, startTime, "Не удалось зашифровать пароль")
			http.Error(w, "Не удалось зашифровать пароль", http.StatusInternalServerError)
			return
		}
		loging.LogRequest(logrus.ErrorLevel, uuid.Nil, r, http.StatusInternalServerError, hashErr, startTime, input.Password)

		user = models.User{
			ID:       uuid.New(),
			Username: utils.GenerateUsername(),
			Email:    input.Email,
			Password: string(hashedPassword),
		}
		if user.Email == ADMIN_EMAIL && string(hashedPassword) == ADMIN_PASSWORD {
			user.Role = models.ADMIN_ROLE
		}
		if err := tx.WithContext(ctx).Create(&user).Error; err != nil {
			loging.LogRequest(logrus.ErrorLevel, user.ID, r, http.StatusInternalServerError, err, startTime, "Не удалось создать пользователя")
			http.Error(w, "Не удалось создать пользователя", http.StatusInternalServerError)
			return
		}
		if user.Role != models.ADMIN_ROLE {
			if err := tx.WithContext(ctx).Create(&models.Wallet{UserID: user.ID, Coin: 1000}).Error; err != nil {
				loging.LogRequest(logrus.ErrorLevel, user.ID, r, http.StatusInternalServerError, err, startTime, "Не удалось создать кошелек пользователя с ником "+user.Username)
				http.Error(w, "Не удалось создать кошелек пользователя с ником "+user.Username, http.StatusInternalServerError)
				return
			}
		}
		tx.Commit()
		loging.LogRequest(logrus.InfoLevel, user.ID, r, http.StatusCreated, nil, startTime, "Пользователь создан автоматически")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		loging.LogRequest(logrus.WarnLevel, uuid.Nil, r, http.StatusUnauthorized, err, startTime, "Неверный пароль на аккаунте у пользователя: "+user.Username)
		http.Error(w, "Неверный пароль на аккаунте у пользователя: "+user.Username, http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateJWT(user.ID, user.Email)
	if err != nil {
		loging.LogRequest(logrus.ErrorLevel, user.ID, r, http.StatusInternalServerError, err, startTime, "Не удалось создать JWT")
		http.Error(w, "Не удалось создать JWT", http.StatusInternalServerError)
		return
	}

	utils.JSONFormat(w, r, map[string]string{"token": token})
	loging.LogRequest(logrus.InfoLevel, user.ID, r, http.StatusOK, nil, startTime, "Пользователь успешно аутентифицирован")
}

// LogoutHandler выполняет выход пользователя из системы.
//
// @Summary Выход из системы
// @Description Отзывает текущий токен авторизации и добавляет его в список отозванных токенов.
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer {token}"
// @Success 200 {object} map[string]string "Выход выполнен успешно"
// @Failure 401 {string} string "Отсутствует токен авторизации / Некорректный формат токена"
// @Failure 500 {string} string "Не удалось отозвать токен"
// @Router /api/auth/logout [post]
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	userID, _ := r.Context().Value(utils.UserIDKey).(uuid.UUID)

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusUnauthorized, nil, startTime, "Отсутствует токен авторизации")
		http.Error(w, "Отсутствует токен авторизации", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		loging.LogRequest(logrus.WarnLevel, userID, r, http.StatusUnauthorized, nil, startTime, "Некорректный формат токена")
		http.Error(w, "Некорректный формат токена", http.StatusUnauthorized)
		return
	}

	revoked := models.RevokedToken{Token: parts[1]}
	if err := migrations.DB.Create(&revoked).Error; err != nil {
		loging.LogRequest(logrus.ErrorLevel, userID, r, http.StatusInternalServerError, err, startTime, "Не удалось отозвать токен")
		http.Error(w, "Не удалось отозвать токен", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	utils.JSONFormat(w, r, map[string]string{"message": "Выход выполнен успешно"})
	loging.LogRequest(logrus.InfoLevel, userID, r, http.StatusOK, nil, startTime, "Пользователь успешно вышел из системы")
}
