package server

import (
	"connectrpc.com/connect"
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
