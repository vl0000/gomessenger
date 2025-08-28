package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/vl0000/gomessenger/server"
)

func main() {
	s := server.New()

	s.Router.Use(middleware.Logger)
	s.Router.Use(middleware.Timeout(5 * time.Second))
	s.Router.Use(httprate.LimitByIP(100, 1*time.Minute))
	s.LoadRoutes()

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
