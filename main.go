package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"time"

	"go.uber.org/zap"
)

func main() {
	lis, err := net.Listen("tcp", "127.0.0.1:8443")
	if err != nil {
		log.Fatalf("error start TCP listener: %v", err)
	}

	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("error create logger: %v", err)
	}

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("request received.", zap.String("method", r.Method), zap.String("path", r.URL.Path))
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Server-Name", "troydai/http-beacon")
			w.WriteHeader(http.StatusOK)
		}),
	}

	chServerStopped := make(chan struct{})
	chSignalTerm := make(chan os.Signal, 1)
	chExit := make(chan int)

	go func() {
		defer close(chServerStopped)
		logger.Info("server started.")
		err := server.ServeTLS(
			lis,
			path.Join(os.Getenv("PWD"), "certs/cert.pem"),
			path.Join(os.Getenv("PWD"), "certs/key.pem"),
		)
		logger.Info("server stopped. an error is always returned.", zap.Error(err))
	}()

	go func() {
		signal.Notify(chSignalTerm, os.Interrupt)
		<-chSignalTerm
		if server != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			logger.Info("shutting down server. wait for at most 10 seconds.")

			err := server.Shutdown(ctx)
			logger.Info("server closed. an error may be returned", zap.Error(err))
		}
		chExit <- 0
		close(chExit)
	}()

	select {
	case <-chServerStopped:
		logger.Info("server stopped. exiting.")
		os.Exit(0)
	case code := <-chExit:
		logger.Info("exit signal received. exiting.")
		os.Exit(code)
	}
}
