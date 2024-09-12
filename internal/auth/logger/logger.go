package logger

import (
	"go.uber.org/zap"
	"log"
)

// MustInit - инициализирует новый logger приложения.
func MustInit(debug bool) (logger *zap.Logger) {
	var err error
	if debug == true {
		logger, err = zap.NewDevelopment()
		if err != nil {
			log.Fatalf("Ошибка инициализации логгера: %v", err)
		}
	} else {
		logger, err = zap.NewProduction()
		if err != nil {
			log.Fatalf("Ошибка инициализации логгера: %v", err)
		}
	}
	return logger
}
