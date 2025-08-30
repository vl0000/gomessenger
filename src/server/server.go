package server

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"

	"connectrpc.com/connect"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth/v5"

	"github.com/vl0000/gomessenger/data"
	messagingv1 "github.com/vl0000/gomessenger/gen/messaging/v1"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const CHANNEL_SIZE int = 32

type MessagingServer struct {
	Addr      string
	Router    *chi.Mux
	Db        *sql.DB
	TokenAuth *jwtauth.JWTAuth
	// Used to communicate with server streams opened with GetDMs()
	Conns map[string]chan *messagingv1.Message
}

func New() *MessagingServer {
	server := MessagingServer{}

	if host, ok := os.LookupEnv("HOST"); ok {
		server.Addr = host
	} else {
		server.Addr = "localhost:3000"
	}

	if _, ok := os.LookupEnv("SECRET_KEY"); !ok {
		log.Fatal("No secret key was set. Check Dockerfile")
	}

	if _, ok := os.LookupEnv("DB_SCHEMA_PATH"); !ok {
		os.Setenv("DB_SCHEMA_PATH", "./data/database.sql")
	}

	db, err := data.SetupTestDatabase(os.Getenv("DB_PATH"))
	if err != nil || db == nil {
		log.Fatalf("Could not setup DB. Error:\n\t%s", err)
	}
	server.Db = db

	server.Router = chi.NewRouter()

	return &server
}

func (s *MessagingServer) Run() error {
	s.TokenAuth = jwtauth.New("HS256", []byte(os.Getenv("SECRET_KEY")), nil)
	s.Conns = make(map[string]chan *messagingv1.Message)

	log.Printf("Starting server in address: %s", s.Addr)
	return http.ListenAndServe(s.Addr, h2c.NewHandler(s.Router, &http2.Server{}))
}

func (s *MessagingServer) Shutdown() {
	log.Println("Shutting down")
	s.Db.Close()
	for _, channel := range s.Conns {
		close(channel)
	}
}

func (s *MessagingServer) SendDirectMessage(
	ctx context.Context,
	req *connect.Request[messagingv1.SendDirectMessageRequest],
) (*connect.Response[messagingv1.SendDirectMessageResponse], error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if err := s.validateSendDirectMessageRequest(req); err != nil {
		return nil, err
	}

	res, err := DoSendDirectMessageWork(s.Db, ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	if channel, ok := s.Conns[req.Msg.Message.Receiver+req.Msg.Message.Sender]; ok {
		channel <- res
	}

	return connect.NewResponse(&messagingv1.SendDirectMessageResponse{Message: res}), nil
}

func (s *MessagingServer) GetDMs(
	ctx context.Context,
	req *connect.Request[messagingv1.GetDMsRequest],
	stream *connect.ServerStream[messagingv1.GetDMsResponse],
) error {

	if err := ctx.Err(); err != nil {
		return err
	}

	err := s.validateGetDMsRequest(req)
	if err != nil {
		return err
	}

	res, err := DoGetDMsWork(s.Db, ctx, req.Msg)
	if err != nil {
		return connect.NewError(connect.CodeUnknown, err)
	}

	if err = stream.Conn().Send(res); err != nil {
		return err
	}

	s.Conns[req.Msg.UserA+req.Msg.UserB] = make(chan *messagingv1.Message, CHANNEL_SIZE)

	for {
		if channel, ok := s.Conns[req.Msg.UserA+req.Msg.UserB]; ok {

			err = stream.Send(&messagingv1.GetDMsResponse{
				Messages: []*messagingv1.Message{<-channel},
			})

			if err != nil {
				return err
			}
		} else {
			break
		}
	}

	return nil
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

	if err := s.validateRegistrationRequest(req); err != nil {
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

	if err := s.validateLoginRequest(req); err != nil {
		return nil, err
	}

	response, err := DoLoginWork(s.Db, s.TokenAuth, ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(response), nil

}

func (s *MessagingServer) GetUserInfo(
	ctx context.Context,
	req *connect.Request[messagingv1.GetUserInfoRequest],
) (*connect.Response[messagingv1.GetUserInfoResponse], error) {

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	err := s.validateGetUserInfo(req)
	if err != nil {
		return nil, err
	}

	response, err := DoGetUserInfoWork(s.Db, ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	return connect.NewResponse(response), nil

}
