package server

import (
	"regexp"
	"time"

	"connectrpc.com/connect"
	messagingv1 "github.com/vl0000/gomessenger/gen/messaging/v1"
)

func (s *MessagingServer) validateLoginRequest(req *connect.Request[messagingv1.LoginRequest]) error {
	if req.Msg.Password == "" || req.Msg.PhoneNumber == "" {
		return connect.NewError(connect.CodeInvalidArgument, &connect.Error{})
	}
	return nil
}

func (s *MessagingServer) validateRegistrationRequest(req *connect.Request[messagingv1.RegisterUserRequest]) error {
	if req.Msg.Password == "" || req.Msg.PhoneNumber == "" || req.Msg.Username == "" {
		return connect.NewError(connect.CodeInvalidArgument, &connect.Error{})
	}
	exists, err := CheckUserExists(s.Db, req.Msg.PhoneNumber)
	if err != nil || exists {
		return connect.NewError(connect.CodeAlreadyExists, err)
	}
	if !regexp.MustCompile(`^\d{3}-\d{3}$`).MatchString(req.Msg.PhoneNumber) {
		return connect.NewError(connect.CodeInvalidArgument, nil)
	}
	return nil
}

func (s *MessagingServer) validateSendDirectMessageRequest(
	req *connect.Request[messagingv1.SendDirectMessageRequest],
) error {
	if req.Msg.Message.Sender == "" || req.Msg.Message.Receiver == "" || req.Msg.Message.Content == "" {
		return connect.NewError(connect.CodeInvalidArgument, nil)
	}
	if req.Msg.Message.Sender == req.Msg.Message.Receiver {
		return connect.NewError(connect.CodeInvalidArgument, nil)
	}

	jwt_str := req.Header().Get("Authorization")
	token, err := s.TokenAuth.Decode(jwt_str)
	if err != nil {
		return connect.NewError(connect.CodeUnauthenticated, err)
	}

	if token.Expiration().Before(time.Now()) {
		return connect.NewError(connect.CodeUnauthenticated, nil)
	}

	if req.Msg.Message.Sender != token.Subject() {
		return connect.NewError(connect.CodeUnauthenticated, nil)
	}

	exists, err := CheckUserExists(s.Db, token.Subject())
	if err != nil || !exists {
		return connect.NewError(connect.CodeUnauthenticated, err)
	}

	return nil
}

func (s *MessagingServer) validateGetDMsRequest(req *connect.Request[messagingv1.GetDMsRequest]) error {

	jwt_str := req.Header().Get("Authorization")
	token, err := s.TokenAuth.Decode(jwt_str)
	if err != nil {
		return connect.NewError(connect.CodeUnauthenticated, err)
	}

	if token.Expiration().Before(time.Now()) {
		return connect.NewError(connect.CodeUnauthenticated, nil)
	}

	exists, err := CheckUserExists(s.Db, token.Subject())
	if err != nil || !exists {
		return connect.NewError(connect.CodeUnauthenticated, err)
	}

	if token.Subject() != req.Msg.UserA {
		return connect.NewError(connect.CodeUnauthenticated, err)
	}

	return nil
}

func (s *MessagingServer) validateGetUserInfo(req *connect.Request[messagingv1.GetUserInfoRequest]) error {

	jwt_str := req.Header().Get("Authorization")
	token, err := s.TokenAuth.Decode(jwt_str)
	if err != nil {
		return connect.NewError(connect.CodeUnauthenticated, err)
	}

	exists, err := CheckUserExists(s.Db, token.Subject())
	if err != nil || !exists {
		return connect.NewError(connect.CodeNotFound, err)
	}

	if token.Expiration().Before(time.Now()) {
		return connect.NewError(connect.CodeUnauthenticated, nil)
	}

	return nil
}
