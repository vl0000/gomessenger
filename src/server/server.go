package server

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"connectrpc.com/connect"
	"github.com/go-chi/chi/v5"
	messagingv1 "github.com/vl0000/gomessenger/gen/messaging/v1"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

type MessagingServer struct {
	Addr   string
	Router *chi.Mux
	db     *sql.DB
}

func (s *MessagingServer) Start() {
	log.Printf("Starting server in address: %s", s.Addr)
	http.ListenAndServe(s.Addr, h2c.NewHandler(s.Router, &http2.Server{}))
}

func (s *MessagingServer) SendDirectMessage(
	ctx context.Context,
	req *connect.Request[messagingv1.SendDirectMessageRequest],
) (*connect.Response[messagingv1.SendDirectMessageResponse], error) {
	log.Println("DM")
	res := connect.NewResponse(&messagingv1.SendDirectMessageResponse{})
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
	return res, nil
}
