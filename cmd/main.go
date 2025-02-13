package main

import (
	"Shop/config"
	"Shop/database/migrations"
	"Shop/database/models"
	"Shop/handlers"
	"Shop/loging"
	"Shop/utils"
	"context"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"os/signal"
	"time"
)

func main() {
	loging.InitLogging()
	config.LoadEnv()
	migrations.InitDB()
	config.InitRedis()

	r := mux.NewRouter()
	loging.Log.Info("Сервер запущен успешно")

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/ping", handlers.PingHandler).Methods("GET")
	apiRouter.HandleFunc("/auth", handlers.AuthHandler).Methods("POST")
	apiRouter.HandleFunc("/auth/logout", handlers.LogoutHandler).Methods("POST")
	apiRouter.HandleFunc("/merch", handlers.ShowMerchHandler).Methods("GET")

	employeeInfoRouter := apiRouter.PathPrefix("/info").Subrouter()
	employeeInfoRouter.Use(utils.AuthMiddleware(models.EMPLOYEE_ROLE))
	employeeInfoRouter.HandleFunc("", handlers.InformationHandler).Methods("GET")

	employeeSendCoinRouter := apiRouter.PathPrefix("/sendCoin").Subrouter()
	employeeSendCoinRouter.Use(utils.AuthMiddleware(models.EMPLOYEE_ROLE))
	employeeSendCoinRouter.HandleFunc("", handlers.SendCoinHandler).Methods("POST")

	employeeBuyItemRouter := apiRouter.PathPrefix("/buy/{item}").Subrouter()
	employeeBuyItemRouter.Use(utils.AuthMiddleware(models.EMPLOYEE_ROLE))
	employeeBuyItemRouter.HandleFunc("", handlers.BuyItemHandler).Methods("GET")

	apiRouter.HandleFunc("", handlers.InformationHandler).Methods("GET")
	adminRouter := apiRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(utils.AuthMiddleware(models.ADMIN_ROLE))
	adminRouter.HandleFunc("/merch", handlers.ShowMerchHandler).Methods("GET")
	adminRouter.HandleFunc("/users", handlers.ShowUserHandler).Methods("GET")
	adminRouter.HandleFunc("/users/{username}", handlers.PutMoneyHandler).Methods("POST")
	adminRouter.HandleFunc("/merch/new", handlers.AddOrChangeMerchHandler).Methods("POST")

	server := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		loging.Log.Info("Сервер успешно запущен на порту: 8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			loging.Log.Fatal("Ошибка сервера:", err)
		}
	}()

	<-stop

	loging.Log.Info("Выключение сервера...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		loging.Log.Fatal("Ошибка при выключении сервера:", err)
	}

	loging.Log.Info("Сервер выключен")
}
