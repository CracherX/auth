package api

import (
	"encoding/json"
	"errors"
	"github.com/CracherX/auth/internal/auth/dto"
	ce "github.com/CracherX/auth/internal/auth/errors"
	"github.com/CracherX/auth/internal/auth/middleware"
	"github.com/CracherX/auth/internal/auth/services"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
)

type TokenEndpoint struct {
	Service *services.TokenService
}

func NewAccessEndpoint(service *services.TokenService) *TokenEndpoint {
	ep := &TokenEndpoint{
		Service: service,
	}
	return ep
}

func (s *TokenEndpoint) Access(w http.ResponseWriter, r *http.Request) {
	var logger = middleware.GetLogger(r.Context())
	var validator = middleware.GetValidator(r.Context())
	vars := mux.Vars(r)
	guid := vars["GUID"]
	reqDat := &dto.AccessRequest{
		GUID: guid,
		IP:   r.Header.Get("X-Forwarded-For"), // Очень сомневаюсь, что наш сервер подразумевает работу с клиентом на прямую
	}
	err := validator.Struct(reqDat)
	if err != nil {
		dto.Error(w, http.StatusBadRequest, "Отправленный формат данных не поддерживается", "Для данного маршрута необходимо указать GUID в URL и IP пользователя в Header запроса")
		logger.Info("Bad Request")
		return
	}
	refTkn, rid, err := s.Service.CreateRefreshToken(reqDat)
	accTkn, err := s.Service.CreateAccessToken(reqDat, rid)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			dto.Error(w, http.StatusUnauthorized, "Пользователь с таким GUID отсутствует")
			logger.Info("Unauthorized. Указан неверный GUID пользователя", zap.String("Указанный IP:", reqDat.IP))
		} else {
			dto.Error(w, http.StatusInternalServerError, "Ошибка на стороне сервера")
			logger.Error("Ошибка создания токенов", zap.String("Ошибка:", err.Error()))
		}
		return
	}
	resDat := &dto.TokenResponse{
		AccessToken:  accTkn,
		RefreshToken: refTkn,
	}
	err = json.NewEncoder(w).Encode(resDat)
	if err != nil {
		dto.Error(w, http.StatusInternalServerError, "Ошибка на стороне сервера")
		logger.Error("Ошибка при сериализации JSON", zap.String("Ошибка:", err.Error()))
		return
	}
}

func (s *TokenEndpoint) Refresh(w http.ResponseWriter, r *http.Request) {
	var logger = middleware.GetLogger(r.Context())
	var validator = middleware.GetValidator(r.Context())
	var reqDat dto.RefreshRequest
	err := json.NewDecoder(r.Body).Decode(&reqDat)
	reqDat.IP = r.Header.Get("X-Forwarded-For")
	err = validator.Struct(reqDat)
	if err != nil {
		dto.Error(w, http.StatusBadRequest, "Отправленный формат данных не поддерживается", "Для данного маршрута необходимо указать пару Access, Refresh-токенов в теле запроса и IP пользователя параметром X-Forwarded-For в Header запроса")
		logger.Info("Bad Request")
		return
	}
	accTkn, refTkn, err := s.Service.RefreshTokens(&reqDat)
	if err != nil {
		if errors.Is(err, ce.InvalidToken) {
			dto.Error(w, http.StatusUnauthorized, "Пара")
			logger.Info("Unauthorized. Пара токенов не валидна", zap.String("Указанный IP:", reqDat.IP))
		} else {
			dto.Error(w, http.StatusInternalServerError, "Ошибка на стороне сервера")
			logger.Error("Ошибка создания токенов", zap.String("Ошибка:", err.Error()))
		}
		return
	}
	resDat := &dto.TokenResponse{
		AccessToken:  accTkn,
		RefreshToken: refTkn,
	}
	err = json.NewEncoder(w).Encode(resDat)
	if err != nil {
		dto.Error(w, http.StatusInternalServerError, "Ошибка на стороне сервера")
		logger.Error("Ошибка при сериализации JSON", zap.String("Ошибка:", err.Error()))
		return
	}
}
