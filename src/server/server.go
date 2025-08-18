package server

import (
	"context"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha512"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"connectrpc.com/connect"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"

	messagingv1 "github.com/vl0000/gomessenger/gen/messaging/v1"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const (
	PBKDF_KEY_LEN int = 32
	PBKDF_ITER    int = 16384
)

type MessagingServer struct {
	Addr      string
	Router    *chi.Mux
	Db        *sql.DB
	TokenAuth *jwtauth.JWTAuth
}

func (s *MessagingServer) Start() {
	s.TokenAuth = jwtauth.New("HS256", []byte(os.Getenv("SECRET_KEY")), nil)
	log.Printf("Starting server in address: %s", s.Addr)
	http.ListenAndServe(s.Addr, h2c.NewHandler(s.Router, &http2.Server{}))
}

func (s *MessagingServer) SendDirectMessage(
	ctx context.Context,
	req *connect.Request[messagingv1.SendDirectMessageRequest],
) (*connect.Response[messagingv1.SendDirectMessageResponse], error) {

	_, err := s.Db.Exec(`INSERT INTO messages (sender, receiver, content, timestamp) VALUES
		(?, ?, ?, datetime('now'));
		`, req.Msg.Msg.Sender, req.Msg.Msg.Receiver, req.Msg.Msg.Receiver, req.Msg.Msg.Content)

	if err != nil || req == nil {
		res := connect.NewResponse(&messagingv1.SendDirectMessageResponse{
			Status: messagingv1.STATUS_STATUS_FAILURE,
		})
		res.Header().Set("Messaging-Version", "v1")
		return res, fmt.Errorf("SendDirectMessage()\n\t%s", err)
	}
	res := connect.NewResponse(&messagingv1.SendDirectMessageResponse{
		Status: messagingv1.STATUS_STATUS_SUCCESS,
	})
	res.Header().Set("Messaging-Version", "v1")
	return res, nil
}

func (s *MessagingServer) GetDMs(
	ctx context.Context,
	req *connect.Request[messagingv1.GetDMsRequest],
) (*connect.Response[messagingv1.GetDMsResponse], error) {
	log.Println("Retrieve DMS")
	res := connect.NewResponse(&messagingv1.GetDMsResponse{})
	res.Header().Set("Messaging-Version", "v1")

	rows, err := s.Db.Query(`SELECT * FROM messages WHERE
			sender IN (?, ?) AND receiver IN (?, ?) AND
			timestamp BETWEEN ? AND datetime('now')
			;`, req.Msg.Sender, req.Msg.Receiver, req.Msg.FromDate)

	if err != nil {
		return res, fmt.Errorf("GetDMs()\n\t%s", err)
	}

	defer rows.Close()

	for rows.Next() {
		var id uint64
		var sender, receiver, content, timestamp string
		err := rows.Scan(&id, &sender, &receiver, &content, &timestamp)

		if err != nil {
			return res, fmt.Errorf("GetDMs()\n\t%s", err)
		}
		res.Msg.Messages = append(
			res.Msg.Messages,
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

func (s *MessagingServer) RegisterUser(
	ctx context.Context,
	req *connect.Request[messagingv1.RegisterUserRequest],
) (
	*connect.Response[messagingv1.RegisterUserResponse],
	error,
) {

	salt := make([]byte, 24)
	rand.Read(salt)

	hashed_password, err := pbkdf2.Key(sha512.New, req.Msg.Password, salt, PBKDF_ITER, PBKDF_KEY_LEN)
	if err != nil {
		return connect.NewResponse(&messagingv1.RegisterUserResponse{
			Status: *messagingv1.STATUS_STATUS_FAILURE.Enum(),
		}), err
	}

	_, err = s.Db.Exec(`INSERT INTO users (
		username,phone_number, password, salt)
		VALUES(?, ?, ?, ?);`,
		req.Msg.Username,
		req.Msg.PhoneNumber,
		hashed_password,
		string(salt),
	)

	if err != nil {
		return connect.NewResponse(&messagingv1.RegisterUserResponse{
			Status: *messagingv1.STATUS_STATUS_FAILURE.Enum(),
		}), err
	}

	_, jwt_str, err := s.TokenAuth.Encode(map[string]interface{}{
		"username":     req.Msg.Username,
		"phone_number": req.Msg.PhoneNumber,
		"iat":          time.Now().Unix(),
		"exp":          time.Now().Add(480 * time.Hour).Unix(),
	})
	if err != nil {
		return connect.NewResponse(&messagingv1.RegisterUserResponse{
			Status: *messagingv1.STATUS_STATUS_FAILURE.Enum(),
		}), err
	}

	return connect.NewResponse(&messagingv1.RegisterUserResponse{
		Status:   *messagingv1.STATUS_STATUS_SUCCESS.Enum(),
		JwtToken: &jwt_str,
	}), nil
}

func (s *MessagingServer) Login(
	ctx context.Context,
	req *connect.Request[messagingv1.LoginRequest],
) (*connect.Response[messagingv1.LoginResponse], error) {

	q, err := s.Db.Query(`
		SELECT * FROM users WHERE phone_number = ? LIMIT 1;`,
		req.Msg.PhoneNumber,
	)
	if err != nil {
		return connect.NewResponse(&messagingv1.LoginResponse{
			Status: *messagingv1.STATUS_STATUS_FAILURE.Enum(),
		}), err
	}

	defer q.Close()

	if q.Next() {
		var stored_password, phone_number, username, salt string

		q.Scan(&username, &phone_number, &stored_password, &salt)
		hashed_password, err := pbkdf2.Key(sha512.New, req.Msg.Password, []byte(salt), PBKDF_ITER, PBKDF_KEY_LEN)

		if err == nil || string(hashed_password) == stored_password {

			_, jwt_str, err := s.TokenAuth.Encode(map[string]interface{}{
				"username":     username,
				"phone_number": phone_number,
				"iat":          time.Now().Unix(),
				"exp":          time.Now().Add(480 * time.Hour).Unix(),
			})
			if err != nil {
				return nil, err
			}
			return connect.NewResponse(&messagingv1.LoginResponse{
				Status:   *messagingv1.STATUS_STATUS_SUCCESS.Enum(),
				JwtToken: &jwt_str,
			}), nil

		} else {
			return nil, err
		}

	}

	return nil, err
}
