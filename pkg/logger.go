package kubernetescontroller

import (
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
)

func Logger(level string, message string) {
	_, fileName, fileLine, ok := runtime.Caller(1)
	if !ok {
		fileName = "unknown"
		fileLine = 0
	}
	fileStrace := fileName + ":" + strconv.Itoa(fileLine)
	logger := logrus.New()
	file, err := os.OpenFile("app.log", os.O_CREATE|os.O_RDWR|os.O_APPEND, 0666)
	if err != nil {
		logger.Fatal(err)
	}
	defer file.Close()
	logger.SetFormatter(&logrus.JSONFormatter{})
	multiWriter := io.MultiWriter(os.Stdout, file)
	logger.SetOutput(multiWriter)
	Level := strings.ToLower(level)
	if Level == "debug" {
		logger.SetLevel(logrus.DebugLevel)
		logger.WithFields(logrus.Fields{
			"file": fileStrace,
		}).Debug(message)
	} else if Level == "info" {
		logger.SetLevel(logrus.InfoLevel)
		logger.WithFields(logrus.Fields{
			"file": fileStrace,
		}).Info(message)
	} else if Level == "warn" {
		logger.SetLevel(logrus.WarnLevel)
		logger.WithFields(logrus.Fields{
			"file": fileStrace,
		}).Warn(message)
	} else if Level == "error" {
		logger.SetLevel(logrus.ErrorLevel)
		logger.WithFields(logrus.Fields{
			"file": fileStrace,
		}).Error(message)
	} else if Level == "panic" {
		logger.SetLevel(logrus.ErrorLevel)
		logger.WithFields(logrus.Fields{
			"file": fileStrace,
		}).Error(message)
	}
}
