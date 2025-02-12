package loging

import (
	"github.com/sirupsen/logrus"
	"io"
	"log"
	"os"
)

var Log = logrus.New()

func InitLogging() {
	if _, err := os.Stat("logs"); os.IsNotExist(err) {
		if err := os.Mkdir("logs", os.ModePerm); err != nil && !os.IsExist(err) {
			log.Fatalf("failed to create logs directory: %v", err)
		}
	}
	file, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		Log.Warn("Не удалось открыть файл логов, используется стандартный вывод")
	} else {
		multiWriter := io.MultiWriter(file, os.Stdout)
		Log.SetOutput(multiWriter)
	}

	Log.SetFormatter(&logrus.JSONFormatter{})
	Log.SetLevel(logrus.InfoLevel)
}
