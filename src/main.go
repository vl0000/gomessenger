package main

import (
	"log"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vl0000/gomessenger/data"
	"github.com/vl0000/gomessenger/gen/messaging/v1/messagingv1connect"
	"github.com/vl0000/gomessenger/server"
)

func main() {
	s := server.MessagingServer{}
	s.Addr = "localhost:3000"
	db, err := data.SetupTestDatabase("./testdb.db")
	if err != nil {
		log.Fatalf("Could not setup DB. Error:\n\t%s", err)
	}
	s.Db = db

	mux := chi.NewMux()
	mux.Use(middleware.Logger)
	mux.Use(middleware.Timeout(5 * time.Second))

	path, handler := messagingv1connect.NewMessagingServiceHandler(&s)

	mux.Handle(path, handler)
	s.Router = mux

	s.Start()

}
