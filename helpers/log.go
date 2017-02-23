package helpers

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/evalphobia/logrus_sentry"
)

var log *logrus.Logger

func createLogger() {
	logrus.SetFormatter(&logrus.TextFormatter{})
	logrus.SetLevel(logrus.DebugLevel)

	log = logrus.New()
	hook, err := logrus_sentry.NewSentryHook(os.Getenv("SENTRY_DSN"), []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.InfoLevel,
	})

	if err == nil {
		log.Hooks.Add(hook)
	}
}

func GetLogger() *logrus.Logger {
	if log == nil {
		createLogger()
	}
	return log
}
