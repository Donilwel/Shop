package main

import (
	"Shop/config"
	"Shop/database/migrations"
	"Shop/handlers"
	"Shop/loging"
	"github.com/gorilla/mux"
	"net/http"
)

func main() {
	config.LoadEnv()
	migrations.InitDB()
	config.InitRedis()

	r := mux.NewRouter()
	loging.InitLogging()
	loging.Log.Info("Сервер запущен успешно")

	apiRouter := r.PathPrefix("/api").Subrouter()
	apiRouter.HandleFunc("/ping", handlers.PingHandler).Methods("GET")

	if err := http.ListenAndServe("localhost:8080", r); err != nil {
		loging.Log.Fatal("Сервер не заработан.")
		return
	}
	loging.Log.Info("Сервер успешно запущен.")
}
