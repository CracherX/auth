package main

import (
	"github.com/CracherX/auth/pkg/auth/app"
	"log"
)

func main() {
	App, err := app.New()
	if err != nil {
		log.Fatalf("Ошибка запуска приложения")
	}
	App.Run()
}
