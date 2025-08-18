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

type MessagingServer struct {
	Addr       string
	Router     *chi.Mux
	Db         *sql.DB
	token_auth *jwtauth.JWTAuth
}

func (s *MessagingServer) Start() {
	s.token_auth = jwtauth.New("HS256", []byte(os.Getenv("SECRET_KEY")), nil)
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

	hashed_password, err := pbkdf2.Key(sha512.New, req.Msg.Password, salt, 16384, 32)
	if err != nil {
		return connect.NewResponse(&messagingv1.RegisterUserResponse{
			Status: *messagingv1.STATUS_STATUS_FAILURE.Enum(),
		}), err
	}

	_, err = s.Db.Query(`INSERT INTO users (
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

	_, jwt_str, err := s.token_auth.Encode(map[string]interface{}{
		"username":     req.Msg.Username,
		"phone_number": req.Msg.PhoneNumber,
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
