package server

import (
	"context"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"errors"

	"github.com/go-chi/jwtauth/v5"
	messagingv1 "github.com/vl0000/gomessenger/gen/messaging/v1"
)

const (
	PBKDF_KEY_LEN int = 32
	PBKDF_ITER    int = 16384
)

func DoRegisterUserWork(
	db *sql.DB,
	token_auth *jwtauth.JWTAuth,
	ctx context.Context,
	msg *messagingv1.RegisterUserRequest,
) (*messagingv1.RegisterUserResponse, error) {

	salt := make([]byte, 24)
	rand.Read(salt)

	hashed_password, err := pbkdf2.Key(sha512.New, msg.Password, salt, PBKDF_ITER, PBKDF_KEY_LEN)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`INSERT INTO users (
		username,phone_number, password, salt)
		VALUES(?, ?, ?, ?);`,
		msg.Username,
		msg.PhoneNumber,
		hashed_password,
		string(salt),
	)
	if err != nil {
		return nil, err
	}

	jwt_str, err := GenJWTString(token_auth, msg.PhoneNumber, msg.Username)
	if err != nil {
		return nil, err
	}

	return &messagingv1.RegisterUserResponse{
		JwtToken: jwt_str,
	}, nil
}

func DoLoginWork(
	db *sql.DB,
	token_auth *jwtauth.JWTAuth,
	ctx context.Context,
	msg *messagingv1.LoginRequest,
) (*messagingv1.LoginResponse, error) {

	q, err := db.Query(`
		SELECT * FROM users WHERE phone_number = ? LIMIT 1;`,
		msg.PhoneNumber,
	)
	if err != nil {
		return nil, err
	}

	defer q.Close()

	if q.Next() {
		var stored_password, phone_number, username, salt string

		q.Scan(&username, &phone_number, &stored_password, &salt)
		hashed_password, err := pbkdf2.Key(sha512.New, msg.Password, []byte(salt), PBKDF_ITER, PBKDF_KEY_LEN)

		if err == nil || string(hashed_password) == stored_password {

			jwt_str, err := GenJWTString(token_auth, phone_number, username)
			if err != nil {
				return nil, err
			}

			return &messagingv1.LoginResponse{
				JwtToken: jwt_str,
			}, nil

		}
	}

	return nil, errors.New("User not found")
}
