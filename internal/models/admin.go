package models

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt"
	"golang.org/x/crypto/bcrypt"
)

type Admin struct {
	ID       string `json:"id"`
	Password string `json:"password"`
}

func GetAdminFromDB() (Admin, error) {
	var admin Admin
	err := db.Get(&admin, "SELECT * FROM admins LIMIT 1")
	return admin, err
}

func (a *Admin) VerifyPassword(password string) error {
	err := bcrypt.CompareHashAndPassword([]byte(a.Password), []byte(password))
	if err != nil {
		return fmt.Errorf("Wrong Password. Unauthorized: %v", err)
	}
	return nil
}

func (a *Admin) GenerateToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"sub": a.ID,
		"exp": time.Now().Add(time.Hour * 24 * 14).Unix(),
	})

	return token.SignedString([]byte(os.Getenv("SECRET_KEY")))
}

func (a *Admin) GeneratePersistentToken() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims{
		"sub": a.ID,
	})

	return token.SignedString([]byte(os.Getenv("SECRET_KEY")))
}
