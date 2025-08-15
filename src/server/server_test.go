package server_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"connectrpc.com/connect"
	"github.com/vl0000/gomessenger/data"
	messagingv1 "github.com/vl0000/gomessenger/gen/messaging/v1"
	"github.com/vl0000/gomessenger/server"
)

func TestServer(t *testing.T) {
	t.Run("Message persists in db", func(t *testing.T) {
		os.Setenv("DB_SCHEMA_PATH", "./../data/database.sql")
		fmt.Println("Start")
		db, err := data.SetupTestDatabase("./testing.db")
		fmt.Println("DB setup")
		if err != nil {
			t.Fatalf("Could not setup testing database\n\t%s", err)
		}
		if db == nil {
			t.Fatal("DB is a nil pointer")
		}

		fmt.Println("Context")
		s := server.MessagingServer{
			Addr: "localhost:3000",
			Db:   db,
		}
		message_req := messagingv1.SendDirectMessageRequest{
			Msg: &messagingv1.Message{
				Sender:   "123-456",
				Receiver: "654-321",
				Content:  "Hello!!!",
			},
		}

		req := connect.NewRequest(&message_req)
		s.SendDirectMessage(context.TODO(), req)
		os.Remove("./testing.db")
	})
}
