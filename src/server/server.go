package server

import (
	"context"
	"database/sql"
	"fmt"
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
	Db     *sql.DB
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

	_, err := s.Db.Exec(`INSERT INTO messages (sender, receiver, content, timestamp) VALUES
		(?, ?, ?, datetime('now'));
		`, req.Msg.Msg.Sender, req.Msg.Msg.Receiver, req.Msg.Msg.Receiver, req.Msg.Msg.Content)

	if err != nil {
		res := connect.NewResponse(&messagingv1.SendDirectMessageResponse{
			Content: messagingv1.STATUS_STATUS_FAILURE,
		})
		res.Header().Set("Messaging-Version", "v1")
		return res, fmt.Errorf("SendDirectMessage()\n\t%s", err)
	}
	res := connect.NewResponse(&messagingv1.SendDirectMessageResponse{
		Content: messagingv1.STATUS_STATUS_SUCCESS,
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
