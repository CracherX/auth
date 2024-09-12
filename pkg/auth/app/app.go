package app

import (
	"fmt"
	"github.com/CracherX/auth/internal/auth/api"
	"github.com/CracherX/auth/internal/auth/config"
	"github.com/CracherX/auth/internal/auth/logger"
	"github.com/CracherX/auth/internal/auth/middleware"
	"github.com/CracherX/auth/internal/auth/route"
	"github.com/CracherX/auth/internal/auth/services"
	"github.com/CracherX/auth/internal/auth/storage/db"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
)

// App структура приложения.
type App struct {
	Config    *config.Config
	Logger    *zap.Logger
	Database  *gorm.DB
	Router    *mux.Router
	Validator *validator.Validate
}

// New конструктор App
func New() (app *App, err error) {
	app = &App{}
	app.Config = config.MustLoad()
	app.Logger = logger.MustInit(app.Config.Server.Debug)
	app.Database, err = db.Connect(app.Config)
	app.Validator = validator.New()
	if err != nil {
		app.Logger.Error("Ошибка подключения к базе данных: ", zap.Error(err))
	}
	app.Router = route.Setup()

	app.Router.Use(middleware.Logging(app.Logger))
	app.Router.Use(middleware.Validate(app.Validator))

	service := services.NewTokenService(app.Database, app.Config)
	endpoint := api.NewAccessEndpoint(service)

	route.Auth(app.Router, endpoint)

	return app, nil
}

// Run запуск приложения.
func (a *App) Run() {
	a.Logger.Info("Запуск приложения", zap.String("Приложение:", a.Config.Server.AppName))
	a.Logger.Debug("Запущен режим отладки для терминала!")
	err := http.ListenAndServe(a.Config.Server.Port, a.Router)
	if err != nil {
		fmt.Println(err)
		a.Logger.Error("Ошибка запуска сервера")
	}
}
