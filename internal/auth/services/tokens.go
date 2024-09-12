package services

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/CracherX/auth/internal/auth/config"
	"github.com/CracherX/auth/internal/auth/dto"
	ce "github.com/CracherX/auth/internal/auth/errors"
	"github.com/CracherX/auth/internal/auth/storage/models"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"io"
	"os"
	"strconv"
	"time"
)

// TokenService структура сервисного слоя, выполняющего работу с бизнес-логикой токенов.
type TokenService struct {
	Config *config.Config
	DB     *gorm.DB
}

// NewTokenService конструктор TokenService.
func NewTokenService(db *gorm.DB, cfg *config.Config) *TokenService {
	return &TokenService{
		Config: cfg,
		DB:     db,
	}
}

// CreateAccessToken создает новый AccessToken формируя Payload из DTO и связывая себя с новым RefreshToken, через rid.
func (ts *TokenService) CreateAccessToken(request *dto.AccessRequest, rid string) (string, error) {
	var user models.Users
	err := ts.DB.Where("guid = ?", request.GUID).First(&user).Error
	if err != nil {
		return "", gorm.ErrRecordNotFound
	}

	claims := jwt.MapClaims{
		"sub": request.GUID,
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Hour).Unix(),
		"ip":  request.IP,
		"rid": rid,
	}

	jwtKey := jwt.NewWithClaims(jwt.SigningMethodHS512, claims)

	key, err := loadKey(ts.Config.Server.SecretPath)
	if err != nil {
		return "", err
	}

	sigJwt, err := jwtKey.SignedString(key)
	if err != nil {
		return "", err
	}

	return sigJwt, nil
}

// CreateRefreshToken создает новый RefreshToken формируя Payload из DTO и записывает его в БД.
func (ts *TokenService) CreateRefreshToken(request *dto.AccessRequest) (tkn string, rid string, err error) {
	var user models.Users
	err = ts.DB.Where("guid = ?", request.GUID).First(&user).Error
	if err != nil {
		return "", "", gorm.ErrRecordNotFound
	}

	bytes := make([]byte, 32)
	_, err = rand.Read(bytes)
	if err != nil {
		return "", "", err
	}
	tkn = base64.StdEncoding.EncodeToString(bytes)

	ht, err := bcrypt.GenerateFromPassword([]byte(tkn), bcrypt.DefaultCost)

	modelToken := models.RefreshTokens{
		Token:     string(ht),
		UserGuid:  request.GUID,
		ExpiresAt: time.Now().Add(24 * time.Hour * 7),
		IP:        request.IP,
	}
	err = ts.DB.Create(&modelToken).Error
	if err != nil {
		return "", "", err
	}
	rid = strconv.Itoa(modelToken.ID)

	return tkn, rid, nil
}

// RefreshTokens выполняет Refresh операцию для пары связанных Access и Refresh токенов.
func (ts *TokenService) RefreshTokens(request *dto.RefreshRequest) (accTkn string, refTkn string, err error) {
	token, err := jwt.Parse(request.AccessToken, func(token *jwt.Token) (interface{}, error) {
		return loadKey(ts.Config.Server.SecretPath)
	})

	if err != nil || !token.Valid {
		return "", "", ce.InvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", ce.InvalidToken
	}

	guid, ok := claims["sub"].(string)
	if !ok {
		return "", "", ce.InvalidToken
	}

	rid, ok := claims["rid"].(string)
	if !ok {
		return "", "", ce.InvalidToken
	}

	var tknMod models.RefreshTokens
	err = ts.DB.Where("id = ? AND user_guid = ?", rid, guid).First(&tknMod).Error
	if err != nil {
		return "", "", ce.InvalidToken
	}

	err = bcrypt.CompareHashAndPassword([]byte(tknMod.Token), []byte(request.RefreshToken))
	if err != nil {
		return "", "", ce.InvalidToken
	}

	if time.Now().After(tknMod.ExpiresAt) {
		return "", "", ce.InvalidToken
	}

	if tknMod.Revoked {
		return "", "", ce.InvalidToken
	}

	dtoAcc := &dto.AccessRequest{
		GUID: guid,
		IP:   request.IP,
	}

	refTkn, rid, err = ts.CreateRefreshToken(dtoAcc)
	if err != nil {
		return "", "", err
	}

	accTkn, err = ts.CreateAccessToken(dtoAcc, rid)
	if err != nil {
		return "", "", err
	}

	err = ts.DB.Model(&tknMod).Updates(models.RefreshTokens{
		Revoked: true,
	}).Error
	if err != nil {
		return "", "", err
	}

	return accTkn, refTkn, nil
}

func loadKey(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	key := string(data)

	bytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}
