package main

import (
	"Shop/config"
	"Shop/database/migrations"
	"Shop/database/models"
	_ "Shop/docs"
	"Shop/handlers"
	"Shop/loging"
	"Shop/utils"
	"context"
	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
	"net/http"
	"os"
	"os/signal"
	"time"
)

// @title Shop API
// @version 1.0
// @description API для магазина с авторизацией, покупками и админ-панелью
// @host localhost:8080
// @BasePath /api
func main() {
	loging.InitLogging()
	config.LoadEnv()
	migrations.InitDB()
	config.InitRedis()

	r := mux.NewRouter()
	loging.Log.Info("Сервер запущен успешно")

	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/ping", handlers.PingHandler).Methods("GET")
	apiRouter.HandleFunc("/auth", handlers.AuthHandler).Methods("POST")
	apiRouter.HandleFunc("/auth/logout", handlers.LogoutHandler).Methods("POST")
	apiRouter.HandleFunc("/merch", handlers.ShowMerchHandler).Methods("GET")
	apiRouter.HandleFunc("/users", handlers.ShowEmployeesHandler).Methods("GET")

	employeeInfoRouter := apiRouter.PathPrefix("/info").Subrouter()
	employeeInfoRouter.Use(utils.AuthMiddleware(models.EMPLOYEE_ROLE))
	employeeInfoRouter.HandleFunc("", handlers.InformationHandler).Methods("GET")

	employeeSendCoinRouter := apiRouter.PathPrefix("/sendCoin").Subrouter()
	employeeSendCoinRouter.Use(utils.AuthMiddleware(models.EMPLOYEE_ROLE))
	employeeSendCoinRouter.HandleFunc("", handlers.SendCoinHandler).Methods("POST")

	employeeBuyItemRouter := apiRouter.PathPrefix("/buy/{item}").Subrouter()
	employeeBuyItemRouter.Use(utils.AuthMiddleware(models.EMPLOYEE_ROLE))
	employeeBuyItemRouter.HandleFunc("", handlers.BuyItemHandler).Methods("GET")

	adminRouter := apiRouter.PathPrefix("/admin").Subrouter()
	adminRouter.Use(utils.AuthMiddleware(models.ADMIN_ROLE))
	adminRouter.HandleFunc("/users", handlers.PutMoneyHandler).Methods("POST")
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
