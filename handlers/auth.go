package handlers

import (
	"Shop/database/migrations"
	"Shop/database/models"
	"Shop/loging"
	"Shop/utils"
	"encoding/json"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
	"time"
)

func AuthHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	var input struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Number   string `json:"number"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		loging.LogRequest(logrus.WarnLevel, uuid.Nil, r, http.StatusBadRequest, err, startTime, "Некорректное тело запроса")
		http.Error(w, "Некорректное тело запроса", http.StatusBadRequest)
		return
	}

	if input.Email == "" {
		input.Email = "none"
	}

	if input.Number == "" {
		input.Number = "0"
	}

	var user models.User
	err := migrations.DB.Where("user = ?", input.Email).First(&user).Error

	if err != nil {
		if input.Name == "" {
			loging.LogRequest(logrus.WarnLevel, uuid.Nil, r, http.StatusBadRequest, nil, startTime, "Имя пользователя обязательно при первой авторизации")
			http.Error(w, "Имя пользователя обязательно при первой авторизации", http.StatusBadRequest)
			return
		}

		hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
		if hashErr != nil {
			loging.LogRequest(logrus.ErrorLevel, uuid.Nil, r, http.StatusInternalServerError, hashErr, startTime, "Не удалось зашифровать пароль")
			http.Error(w, "Не удалось зашифровать пароль", http.StatusInternalServerError)
			return
		}

		user = models.User{
			ID:          uuid.New(),
			Name:        input.Name,
			Email:       input.Email,
			PhoneNumber: input.Number,
			Password:    string(hashedPassword),
		}

		if createErr := migrations.DB.Create(&user).Error; createErr != nil {
			loging.LogRequest(logrus.ErrorLevel, user.ID, r, http.StatusInternalServerError, createErr, startTime, "Не удалось создать пользователя")
			http.Error(w, "Не удалось создать пользователя", http.StatusInternalServerError)
			return
		}

		loging.LogRequest(logrus.InfoLevel, user.ID, r, http.StatusCreated, nil, startTime, "Пользователь создан автоматически")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		loging.LogRequest(logrus.WarnLevel, uuid.Nil, r, http.StatusUnauthorized, err, startTime, "Неверный пароль на аккаунте у пользователя: "+input.Name)
		http.Error(w, "Неверный email или пароль", http.StatusUnauthorized)
		return
	}

	token, err := utils.GenerateJWT(user.ID, user.Name)
	if err != nil {
		loging.LogRequest(logrus.ErrorLevel, user.ID, r, http.StatusInternalServerError, err, startTime, "Не удалось создать JWT")
		http.Error(w, "Не удалось создать JWT", http.StatusInternalServerError)
		return
	}

	utils.JSONFormat(w, r, map[string]string{"token": token})
	loging.LogRequest(logrus.InfoLevel, user.ID, r, http.StatusOK, nil, startTime, "Пользователь успешно аутентифицирован")
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	userID, _ := r.Context().Value("userID").(uuid.UUID)

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
