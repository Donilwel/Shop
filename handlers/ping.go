package handlers

import (
	"Shop/loging"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func PingHandler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte("pong")); err != nil {
		loging.LogRequest(logrus.WarnLevel, uuid.Nil, r, http.StatusBadRequest, nil, startTime, "Failed to write ping response")
		return
	}
	loging.LogRequest(logrus.InfoLevel, uuid.Nil, r, http.StatusOK, nil, startTime, "Ping request complete")
}
