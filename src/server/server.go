package server

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"

	messagingv1 "github.com/vl0000/gomessenger/gen/messaging/v1"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
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
			;`, req.Msg.Sender, req.Msg.Receiver, req.Msg.Receiver, req.Msg.Sender, req.Msg.FromDate)

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

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := ValidateRegistrationRequest(req.Msg); err != nil {
		return nil, err
	}

	response, err := DoRegisterUserWork(s.Db, s.TokenAuth, ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(response), nil

}

func (s *MessagingServer) Login(
	ctx context.Context,
	req *connect.Request[messagingv1.LoginRequest],
) (*connect.Response[messagingv1.LoginResponse], error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := ValidateLoginRequest(req.Msg); err != nil {
		return nil, err
	}

	response, err := DoLoginWork(s.Db, s.TokenAuth, ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(response), nil

}
