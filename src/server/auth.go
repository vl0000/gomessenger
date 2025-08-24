package server

import (
	"database/sql"
	"time"

	"github.com/go-chi/jwtauth/v5"
)

const JWT_DURATION time.Duration = 480 * time.Hour

func CheckUserExists(db *sql.DB, phone_number string) (bool, error) {
	q, err := db.Query(`
		SELECT * FROM users WHERE phone_number = ? LIMIT 1;`,
		phone_number,
	)
	if err != nil {
		return false, err
	}
	defer q.Close()
	if !q.Next() {
		return false, nil
	}

	return true, nil
}

func GenJWTString(
	token_auth *jwtauth.JWTAuth,
	phone_number string,
	username string,
) (string, error) {

	_, jwt_str, err := token_auth.Encode(map[string]interface{}{
		"username": username,
		"sub":      phone_number,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(JWT_DURATION).Unix(),
	})
	if err != nil {
		return "", err
	}
	return jwt_str, nil
}
