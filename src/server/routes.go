package server

import (
	"log"
	"net/http"
	"os"

	"github.com/vl0000/gomessenger/gen/messaging/v1/messagingv1connect"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func ServeHTML(path string) http.HandlerFunc {
	if page, err := os.ReadFile(path); err != nil {
		log.Printf("%s Not found\n", path)
	} else {

		return func(w http.ResponseWriter, r *http.Request) {
			_, err = w.Write(page)
			if err != nil {
				log.Println(err)
			}
		}
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		http.Error(w, "Not Found", 404)
	}
}

func (s *MessagingServer) LoadRoutes() {

	// Loads static files for svelte apps
	s.Router.Handle("/*", http.FileServer(http.Dir("./public/static/")))

	// Loads the paths for the messaging service
	path, handler := messagingv1connect.NewMessagingServiceHandler(s)
	s.Router.Handle(path+"*", h2c.NewHandler(handler, &http2.Server{}))

	s.Router.Get("/", ServeHTML("./public/login.html"))
	s.Router.Get("/signup", ServeHTML("./public/signup.html"))
	s.Router.Get("/chat", ServeHTML("./public/chat.html"))
}
