package server

import (
	"context"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"errors"

	"connectrpc.com/connect"
	"github.com/go-chi/jwtauth/v5"
	messagingv1 "github.com/vl0000/gomessenger/gen/messaging/v1"
)

const (
	PBKDF_KEY_LEN int = 32
	PBKDF_ITER    int = 210000
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

		q.Scan(&phone_number, &username, &stored_password, &salt)
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

func DoSendDirectMessageWork(
	db *sql.DB,
	ctx context.Context,
	msg *messagingv1.SendDirectMessageRequest,
) (*messagingv1.SendDirectMessageResponse, error) {

	_, err := db.Exec(`INSERT INTO messages (sender, receiver, content, timestamp) VALUES
		(?, ?, ?, datetime('now'));
		`, msg.Msg.Sender, msg.Msg.Receiver, msg.Msg.Content)

	if err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	return &messagingv1.SendDirectMessageResponse{}, nil
}

func DoGetDMsWork(
	db *sql.DB,
	ctx context.Context,
	msg *messagingv1.GetDMsRequest,
) (*messagingv1.GetDMsResponse, error) {
	res := &messagingv1.GetDMsResponse{}

	rows, err := db.Query(`SELECT * FROM messages WHERE
			sender IN (?, ?) AND receiver IN (?, ?) AND
			timestamp BETWEEN ? AND datetime('now')
			;`, msg.Sender, msg.Receiver, msg.Receiver, msg.Sender, msg.FromDate)

	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	defer rows.Close()

	for rows.Next() {
		var id uint64
		var sender, receiver, content, timestamp string
		err := rows.Scan(&id, &sender, &receiver, &content, &timestamp)

		if err != nil {
			return res, connect.NewError(connect.CodeUnknown, err)
		}

		res.Messages = append(
			res.Messages,
			&messagingv1.Message{
				Id:        &id,
				Sender:    sender,
				Receiver:  receiver,
				Content:   content,
				Timestamp: &timestamp,
			})
	}
	return res, nil
}

func DoGetUserInfoWork(
	db *sql.DB,
	ctx context.Context,
	req *messagingv1.GetUserInfoRequest,
) (*messagingv1.GetUserInfoResponse, error) {

	q, err := db.Query(`
		SELECT * FROM users WHERE phone_number = ? LIMIT 1;`,
		req.PhoneNumber,
	)
	if err != nil {
		return nil, err
	}

	defer q.Close()

	if q.Next() {
		var phone_number, username string
		err = q.Scan(&phone_number, &username, nil, nil)
		// rows.Scan will return an error if you do not give it a pointer,
		// but this doesn't stop it from retrieving the necessary data.
		// This means that the error is only relevant these are empty
		if phone_number == "" || username == "" {
			return nil, connect.NewError(connect.CodeUnknown, err)
		}

		return &messagingv1.GetUserInfoResponse{
			PhoneNumber: phone_number,
			Username:    username,
		}, nil
	}

	return nil, connect.NewError(connect.CodeNotFound, nil)
}
