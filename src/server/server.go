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

func (s *MessagingServer) Run() error {
	s.TokenAuth = jwtauth.New("HS256", []byte(os.Getenv("SECRET_KEY")), nil)
	log.Printf("Starting server in address: %s", s.Addr)
	return http.ListenAndServe(s.Addr, h2c.NewHandler(s.Router, &http2.Server{}))
}

func (s *MessagingServer) Shutdown() {
	log.Println("Shutting down")
	s.Db.Close()
}

func (s *MessagingServer) SendDirectMessage(
	ctx context.Context,
	req *connect.Request[messagingv1.SendDirectMessageRequest],
) (*connect.Response[messagingv1.SendDirectMessageResponse], error) {

	// Verify JWT
	jwt_str := req.Header().Get("Authorization")
	token, err := s.TokenAuth.Decode(jwt_str)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	exists, err := CheckUserExists(s.Db, token.Subject())
	if err != nil || !exists {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	if err := ValidateSendDirectMessageRequest(req.Msg, token); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, err)
	}

	res, err := DoSendDirectMessageWork(s.Db, ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	return connect.NewResponse(res), nil
}

func (s *MessagingServer) GetDMs(
	ctx context.Context,
	req *connect.Request[messagingv1.GetDMsRequest],
) (*connect.Response[messagingv1.GetDMsResponse], error) {
	// Verify JWT
	jwt_str := req.Header().Get("Authorization")
	token, err := s.TokenAuth.Decode(jwt_str)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	exists, err := CheckUserExists(s.Db, token.Subject())
	if err != nil || !exists {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	res, err := DoGetDMsWork(s.Db, ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}
	return connect.NewResponse(res), nil
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

func (s *MessagingServer) GetUserInfo(
	ctx context.Context,
	req *connect.Request[messagingv1.GetUserInfoRequest],
) (*connect.Response[messagingv1.GetUserInfoResponse], error) {
	jwt_str := req.Header().Get("Authorization")
	token, err := s.TokenAuth.Decode(jwt_str)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	exists, err := CheckUserExists(s.Db, token.Subject())
	if err != nil || !exists {
		return nil, connect.NewError(connect.CodeUnauthenticated, err)
	}

	response, err := DoGetUserInfoWork(s.Db, ctx, req.Msg)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnknown, err)
	}

	return connect.NewResponse(response), nil

}
