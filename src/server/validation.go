package server

import (
	"connectrpc.com/connect"
	"github.com/lestrrat-go/jwx/v2/jwt"
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
	return nil
}

func (s *MessagingServer) validateSendDirectMessageRequest(
	req *connect.Request[messagingv1.SendDirectMessageRequest],
	token jwt.Token,
) error {
	if req.Msg.Msg.Sender == "" || req.Msg.Msg.Receiver == "" || req.Msg.Msg.Content == "" {
		return connect.NewError(connect.CodeInvalidArgument, nil)
	}
	if req.Msg.Msg.Sender == req.Msg.Msg.Receiver {
		return connect.NewError(connect.CodeInvalidArgument, nil)
	}
	if req.Msg.Msg.Sender != token.Subject() {
		return connect.NewError(connect.CodeUnauthenticated, nil)
	}
	return nil
}
