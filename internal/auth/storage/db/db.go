package db

import (
	"fmt"
	"github.com/CracherX/auth/internal/auth/config"
	"github.com/CracherX/auth/internal/auth/storage/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"strings"
)

// Connect создает новое подключение к БД из параметров конфигурации.
func Connect(cfg *config.Config) (*gorm.DB, error) {
	dsn := parseConfigDSN(cfg)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&models.Users{}, &models.RefreshTokens{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func parseConfigDSN(cfg *config.Config) string {
	params := map[string]string{
		"host":     cfg.Database.Host,
		"port":     cfg.Database.Port,
		"user":     cfg.Database.User,
		"password": cfg.Database.Password,
		"dbname":   cfg.Database.Name,
		"sslmode":  cfg.Database.SslMode,
	}
	var dsnParts []string
	for key, value := range params {
		dsnParts = append(dsnParts, fmt.Sprintf("%s=%s", key, value))
	}
	return strings.Join(dsnParts, " ")
}
