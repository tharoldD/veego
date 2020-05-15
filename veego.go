package veego

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	h "github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type Server struct {
	Config     *AppConfig
	BaseRouter *mux.Router
}

func NewServer(config *AppConfig, baseRouter *mux.Router) *Server {
	server := &Server{
		Config:     config,
		BaseRouter: baseRouter,
	}
	return server
}

func (s *Server) Run() error {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		oscall := <-c
		log.Printf("system call:%+v", oscall)
		cancel()
	}()
	fmt.Printf("Appliction booting on %s:%s...\n", s.Config.Host, s.Config.Port)
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%s", s.Config.Host, s.Config.Port),
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler: h.CORS(h.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization", "Access-Control-Allow-Origin"}),
			h.AllowedMethods([]string{"GET"}),
			h.AllowedOrigins([]string{"*"}))(s.BaseRouter),
	}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("serve failed: %v", err.Error())
		}
	}()
	fmt.Printf("Appliction running")
	<-ctx.Done()
	fmt.Printf("Appliction stopped")
	ctxShutDown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()
	if err := server.Shutdown(ctxShutDown); err != nil {
		log.Fatalf("Application Shutdown Failed:%+s", err)
	}
	return nil
}
