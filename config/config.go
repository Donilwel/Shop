package config

import (
	"github.com/joho/godotenv"
	"log"
)

func LoadEnv() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Ошибка загрузки .env файла")
		return
	}
}
