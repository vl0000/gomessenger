package server

import (
	"connectrpc.com/connect"
	"github.com/lestrrat-go/jwx/v2/jwt"
	messagingv1 "github.com/vl0000/gomessenger/gen/messaging/v1"
)

func ValidateLoginRequest(msg *messagingv1.LoginRequest) error {
	if msg.Password == "" || msg.PhoneNumber == "" {
		return connect.NewError(connect.CodeInvalidArgument, &connect.Error{})
	}
	return nil
}

func ValidateRegistrationRequest(msg *messagingv1.RegisterUserRequest) error {
	if msg.Password == "" || msg.PhoneNumber == "" || msg.Username == "" {
		return connect.NewError(connect.CodeInvalidArgument, &connect.Error{})
	}
	return nil
}

func ValidateSendDirectMessageRequest(
	msg *messagingv1.SendDirectMessageRequest,
	token jwt.Token,
) error {
	if msg.Msg.Sender == "" || msg.Msg.Receiver == "" || msg.Msg.Content == "" {
		return connect.NewError(connect.CodeInvalidArgument, nil)
	}
	if msg.Msg.Sender == msg.Msg.Receiver {
		return connect.NewError(connect.CodeInvalidArgument, nil)
	}
	if msg.Msg.Sender != token.Subject() {
		return connect.NewError(connect.CodeUnauthenticated, nil)
	}
	return nil
}
