package helpers

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
)

func ValidateToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("There was an error")
		}

		return []byte(os.Getenv("JWT_SECRET")), nil
	})

	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		expired := int64(claims["exp"].(float64))

		if time.Now().Unix() > expired {
			return "", fmt.Errorf("Token expired at %v, now is %v", time.Unix(expired, 0), time.Now())
		}

		return claims["sub"].(string), nil
	} else {
		return "", fmt.Errorf("Invalid Token")
	}
}
