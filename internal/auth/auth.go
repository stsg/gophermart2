package auth

import (
	"time"

	"github.com/stsg/gophermart2/internal/config"
	"github.com/stsg/gophermart2/internal/models"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

const TokenExpTime = 15 * time.Minute

func GenerateToken(user models.User) (signedToken string, err error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uid": user.ID,
		"exp": time.Now().Add(TokenExpTime).Unix(),
	})
	signedToken, err = token.SignedString([]byte(config.Get().SecretToken))
	return
}

func GetUIDFromValidToken(signedToken string) (string, error) {
	token, err := jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Get().SecretToken), nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", models.ErrInvalidBearerToken
	}
	return claims["uid"].(string), nil
}

func Authenticate(dbUser models.User, reqUser models.User) bool {
	if err := bcrypt.CompareHashAndPassword([]byte(dbUser.Password), []byte(reqUser.Password)); err != nil {
		return false
	}
	return true
}
