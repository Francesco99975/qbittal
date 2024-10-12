package models

import (
	"log"
	"os"

	"crypto/rand"
	"math/big"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

func generateRandomPassword(length int) (string, error) {

	availableChars := "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghkmnpqrstuvwxyz23456789"
	password := make([]byte, length)

	for i := 0; i < length; i++ {
		idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(availableChars))))
		if err != nil {
			return "", err
		}
		password[i] = availableChars[idx.Int64()]
	}

	return string(password), nil
}

var db *sqlx.DB

func Setup(dsn string) string {
	var err error
	db, err = sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Fatalln(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalln(err)
	}

	schema, err := os.ReadFile("sql/init.sql")
	if err != nil {
		log.Fatalln(err)
	}

	db.MustExec(string(schema))

	var count int

	rows, err := db.Query("SELECT COUNT(*) FROM admins;")

	if err != nil {
		log.Fatalln(err)
	}

	for rows.Next() {
		err = rows.Scan(&count)
		if err != nil {
			log.Fatalln(err)
		}
	}

	password := ""

	if count == 0 {
		id := uuid.New().String()
		password, err = generateRandomPassword(10)
		if err != nil {
			log.Fatalln(err)
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), 12)
		if err != nil {
			log.Fatalln(err)
		}

		statement := "INSERT INTO admins(id, password) VALUES($1, $2);"

		_, err = db.Exec(statement, id, hashedPassword)

		if err != nil {
			log.Fatalln(err)
		}
	}

	return password
}
