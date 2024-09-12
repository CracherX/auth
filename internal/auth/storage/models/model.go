package models

import (
	"time"
)

// Users модель пользователей БД.
type Users struct {
	Guid  string `gorm:"type:uuid;primaryKey"`
	Email string `gorm:"type:varchar(60);uniqueIndex;not null"`
}

// RefreshTokens модель Refresh токенов БД.
type RefreshTokens struct {
	ID        int       `gorm:"primaryKey;autoIncrement"`
	Token     string    `gorm:"type:varchar;uniqueIndex;not null"`
	UserGuid  string    `gorm:"type:uuid;not null"`
	ExpiresAt time.Time `gorm:"type:timestamp;not null"`
	CreatedAt time.Time `gorm:"type:timestamp;autoCreateTime;not null"`
	UpdatedAt time.Time `gorm:"type:timestamp;autoUpdateTime;not null"`
	IP        string    `gorm:"type:varchar(45);not null"`
	Revoked   bool      `gorm:"type:boolean;not null;default:false"`

	User Users `gorm:"foreignKey:UserGuid"`
}
