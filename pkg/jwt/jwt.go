package core_jwt

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/poponyas/AuthService/internal/core/domain"
	"strconv"
	"time"
)

func NewToken(user domain.User, app domain.App, duration time.Duration, key []byte) (string, error) {
	now := time.Now()
	claims := jwt.MapClaims{"sub": strconv.FormatInt(user.ID, 10), "uid": user.ID, "aud": app.Name, "iss": "service-auth", "iat": now.Unix(), "nbf": now.Unix(), "exp": now.Add(duration).Unix(), "app_id": app.ID}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(key)
}
