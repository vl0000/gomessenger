package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/vl0000/gomessenger/gen/messaging/v1/messagingv1connect"
	"github.com/vl0000/gomessenger/server"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {
	s := server.New()

	path, handler := messagingv1connect.NewMessagingServiceHandler(s)

	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.Timeout(5 * time.Second))
	s.Router.Use(httprate.LimitByIP(100, 1*time.Minute))

	s.Router.Handle(path+"*", h2c.NewHandler(handler, &http2.Server{}))
	s.Router.Handle("/*", http.FileServer(http.Dir("./public/_app/")))

	go func() {
		err := s.Run()
		if err != nil {
			log.Println(err)
		}
	}()

	shutdown, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()

	<-shutdown.Done()

	s.Shutdown()
}
