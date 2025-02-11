package main

import (
	"Shop/config"
	"Shop/database/migrations"
	"Shop/database/models"
	"Shop/handlers"
	"Shop/loging"
	"Shop/utils"
	"github.com/gorilla/mux"
	"net/http"
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
	apiRouter.HandleFunc("/merch", handlers.ShowMerchHandler).Methods("GET")

	adminRouter := r.PathPrefix("/admin").Subrouter()
	adminRouter.Use(utils.AuthMiddleware(models.ADMIN_ROLE))
	adminRouter.HandleFunc("/merch", handlers.ShowMerchHandler).Methods("GET")

	if err := http.ListenAndServe(":8080", r); err != nil {
		loging.Log.Fatal("Сервер не заработал.")
	}
	loging.Log.Info("Сервер успешно запущен на порту: ", 8080)
}
