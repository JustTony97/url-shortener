package handler

import (
	"net/http"
	"time"

	"github.com/JustTony97/url-shortener.git/internal/config"
	"github.com/JustTony97/url-shortener.git/internal/logger"
	"github.com/JustTony97/url-shortener.git/internal/model"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type AuthHandler struct {
	cfg config.Config
}

func NewAuthHandler(cfg config.Config) *AuthHandler {
	return &AuthHandler{cfg: cfg}
}

func (au *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, model.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(au.cfg.JWTExp)),
		},
		UserID: uuid.NewString(),
	})

	tokenString, err := token.SignedString([]byte(au.cfg.JWTSecret))
	if err != nil {
		logger.Log.Error(err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	cookie := http.Cookie{
		Name:     "Authorization",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(au.cfg.JWTExp),
	}
	http.SetCookie(w, &cookie)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}
