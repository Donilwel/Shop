package loging

import (
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
)

func LogRequest(level logrus.Level, userID uuid.UUID, r *http.Request, status int, err error, startTime time.Time, message string) {
	fields := logrus.Fields{
		"userID": userID,
		"path":   r.URL.Path,
		"method": r.Method,
		"status": status,
		"time":   time.Since(startTime).String(),
	}

	if err != nil {
		fields["error"] = err.Error()
	}

	entry := Log.WithFields(fields)

	switch level {
	case logrus.InfoLevel:
		entry.Info(message)
	case logrus.WarnLevel:
		entry.Warn(message)
	case logrus.ErrorLevel:
		entry.Error(message)
	case logrus.FatalLevel:
		entry.Fatal(message)
	}
}
