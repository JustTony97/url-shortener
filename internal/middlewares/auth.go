package middlewares

import (
	"context"
	"net/http"
	"time"

	"github.com/JustTony97/url-shortener.git/internal/config"
	"github.com/JustTony97/url-shortener.git/internal/model"
	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type key string

const userIDKey key = "userID"

type AuthMiddleware struct {
	cfg config.Config
}

func NewAuthMiddleware(cfg config.Config) *AuthMiddleware {
	return &AuthMiddleware{cfg: cfg}
}

func (am *AuthMiddleware) WithAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := am.getUserIDFromCookie(r)

		if uid == "" {
			uid = uuid.New().String()
			claims := &model.Claims{
				UserID: uid,
				RegisteredClaims: jwt.RegisteredClaims{
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(am.cfg.JWTExp)),
				},
			}
			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
			if tokenString, err := token.SignedString([]byte(am.cfg.JWTSecret)); err == nil {
				http.SetCookie(w, &http.Cookie{
					Name:   "Authorization",
					Value:  tokenString,
					Path:   "/",
					MaxAge: int(am.cfg.JWTExp.Seconds()),
				})
			}
		}

		ctx := context.WithValue(r.Context(), userIDKey, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (am *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		uid := am.getUserIDFromCookie(r)
		if uid == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userIDKey, uid)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (am *AuthMiddleware) getUserIDFromCookie(r *http.Request) string {
	cookie, err := r.Cookie("Authorization")
	if err != nil {
		return ""
	}

	claims := &model.Claims{}
	token, err := jwt.ParseWithClaims(cookie.Value, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(am.cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return ""
	}

	return claims.UserID
}

func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}
