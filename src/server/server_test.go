package server_test

import (
	"context"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha512"
	"fmt"
	"os"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/go-chi/jwtauth/v5"
	"github.com/vl0000/gomessenger/data"
	messagingv1 "github.com/vl0000/gomessenger/gen/messaging/v1"
	"github.com/vl0000/gomessenger/server"
)

func newTestingServer() (*server.MessagingServer, error) {

	os.Remove("./testing.db")
	os.Setenv("DB_SCHEMA_PATH", "./../data/database.sql")
	db, err := data.SetupTestDatabase("./testing.db")
	if err != nil || db == nil {
		return nil, fmt.Errorf("DATABASE ERROR >>%s", err)
	}
	return &server.MessagingServer{
		Addr:      "localhost:3000",
		Db:        db,
		TokenAuth: jwtauth.New("HS256", []byte(os.Getenv("SECRET_KEY")), nil),
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
		// SETUP
		s, err := newTestingServer()
		if err != nil {
			t.Fatal(err)
		}

		// END SETUP
		req := connect.NewRequest(&message_req)
		// This bypasses the JWT authentication
		server.DoSendDirectMessageWork(s.Db, context.TODO(), req.Msg)
		q, err := s.Db.Query("SELECT * FROM messages WHERE sender = '123-456';")
		if err != nil {
			t.Fatal(err)
		}
		if !q.Next() {
			t.Fatal("Message not found in DB")
		}
		os.Remove("./testing.db")
	})

	t.Run("Messages can be retrieved from DB", func(t *testing.T) {

		// SETUP
		s, err := newTestingServer()
		if err != nil {
			t.Fatal(err)
		}
		req := connect.NewRequest(&messagingv1.GetDMsRequest{
			Sender:   "123-456",
			Receiver: "654-321",
			FromDate: time.Now().Add(-24 * time.Hour).Format(time.DateTime),
		})
		_, err = s.Db.Exec(`INSERT INTO messages (sender, receiver, content, timestamp) VALUES
			(?, ?, ?, datetime('now'));
			`, req.Msg.Sender, req.Msg.Receiver, "Hello, World!")
		if err != nil {
			t.Fatal(err)
		}
		// END SETUP

		// Bypass JWT
		res, err := server.DoGetDMsWork(s.Db, context.TODO(), req.Msg)
		if err != nil {
			t.Fatal(err)
		} else if len(res.GetMessages()) == 0 {
			t.Fatalf("No messages returned")
		}
	})

	t.Run("Login requests", func(t *testing.T) {

		// A SECRET KEY MUST BE SET FOR THIS TEST TO RUN CORRECTLY!!!
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

	t.Run("Registers user", func(t *testing.T) {
		// SETUP
		// A SECRET_KEY MUST BE SET FOR THIS TEST TO RUN CORRECTLY!!!
		s, err := newTestingServer()
		if err != nil {
			t.Fatal(err)
		}
		req := connect.NewRequest(&messagingv1.RegisterUserRequest{
			Username:    "John Doe",
			PhoneNumber: "123-456",
			Password:    "12345678",
		})
		// END SETUP

		_, err = s.RegisterUser(context.TODO(), req)
		if err != nil {
			t.Fatal(err)
		}
		os.Remove("./testing.db")
	})

}
