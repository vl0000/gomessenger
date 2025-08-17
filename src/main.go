package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/vl0000/gomessenger/data"
	"github.com/vl0000/gomessenger/gen/messaging/v1/messagingv1connect"
	"github.com/vl0000/gomessenger/server"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	s := server.MessagingServer{}
	s.Addr = "localhost:3000"
	os.Setenv("DB_SCHEMA_PATH", "./data/database.sql")
	os.Setenv("STATIC_FOLDER", "./public/static")
	db, err := data.SetupTestDatabase("./testdb.db")
	if err != nil || db == nil {
		log.Fatalf("Could not setup DB. Error:\n\t%s", err)
	}
	s.Db = db

	path, handler := messagingv1connect.NewMessagingServiceHandler(&s)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Timeout(5 * time.Second))

	r.Handle(path+"*", h2c.NewHandler(handler, &http2.Server{}))

	r.Handle("/*", http.FileServer(http.Dir(os.Getenv("STATIC_FOLDER"))))

	s.Router = r

	s.Start()

}
