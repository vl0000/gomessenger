package server_test

import (
	"context"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha512"
	"os"
	"testing"

	"connectrpc.com/connect"
	"github.com/vl0000/gomessenger/data"
	messagingv1 "github.com/vl0000/gomessenger/gen/messaging/v1"
	"github.com/vl0000/gomessenger/server"
)

func newTestingServer() (*server.MessagingServer, error) {

	os.Setenv("DB_SCHEMA_PATH", "./../data/database.sql")
	db, err := data.SetupTestDatabase("./testing.db")
	if err != nil {
		return nil, err
	}
	return &server.MessagingServer{
		Addr: "localhost:3000",
		Db:   db,
	}, nil
}

func TestServer(t *testing.T) {
	t.Run("Message persists in db", func(t *testing.T) {
		message_req := messagingv1.SendDirectMessageRequest{
			Msg: &messagingv1.Message{
				Sender:   "123-456",
				Receiver: "654-321",
				Content:  "Hello!!!",
			},
		}
		s, err := newTestingServer()
		if err != nil {
			t.Fatal(err)
		}

		req := connect.NewRequest(&message_req)
		s.SendDirectMessage(context.TODO(), req)
		q, err := s.Db.Query("SELECT * FROM messages WHERE sender = '123-456';")
		if err != nil {
			t.Fatal(err)
		}
		if !q.Next() {
			t.Fatal("Message not found in DB")
		}
		os.Remove("./testing.db")
	})

	t.Run("Login requests", func(t *testing.T) {

		const (
			PBKDF_KEY_LEN int = 32
			PBKDF_ITER    int = 16384
		)

		req := messagingv1.LoginRequest{
			PhoneNumber: "123-456",
			Password:    "123456",
		}

		// SETUP
		s, err := newTestingServer()
		if err != nil {
			t.Fatal(err)
		}

		salt := make([]byte, 24)
		rand.Read(salt)

		hashed_password, err := pbkdf2.Key(sha512.New, req.PhoneNumber, salt, PBKDF_ITER, PBKDF_KEY_LEN)

		_, err = s.Db.Exec(`INSERT INTO users (
			username,phone_number, password, salt)
			VALUES(?, ?, ?, ?);`,
			"John Doe",
			req.PhoneNumber,
			hashed_password,
			string(salt),
		)
		// END SETUP

		_, err = s.Login(context.TODO(), connect.NewRequest(&req))
		if err != nil {
			t.Fatal(err)
		}
		os.Remove("./testing.db")
	})

}
