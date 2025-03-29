package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"time"

	"github.com/caarlos0/env/v11"
)

type options struct {
	TLSOption string `env:"BEACON_TLS_OPTION" envDefault:"tls" enums:"tls,plaintexth2,plaintexth1"` // tls: TLS server, plaintexth2: unencrypted HTTP/2, plaintexth1: unencrypted HTTP/1.1
}

func main() {
	var opts options
	if err := env.Parse(&opts); err != nil {
		// crash out early, the logs level will need to be decided by environment variable as well
		log.Fatal("failed to parse environment variables", "error", err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	logger.Info("starting server", "opts", opts)

	lc := &net.ListenConfig{KeepAlive: -1}
	lis, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:8443")
	if err != nil {
		logger.Error("error start TCP listener", "error", err)
		os.Exit(1)
	}

	tlsConfig, err := createTLSConfig(opts)
	if err != nil {
		logger.Error("error creating TLS config", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("request received", "method", r.Method, "path", r.URL.Path)
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("Server-Name", "troydai/http-beacon")
			w.WriteHeader(http.StatusOK)
		}),
		TLSConfig: tlsConfig,
	}

	chServerStopped := make(chan struct{})
	chSignalTerm := make(chan os.Signal, 1)
	chExit := make(chan int)

	go func() {
		defer close(chServerStopped)
		logger.Info("server started")

		var err error
		if opts.TLSOption == "tls" {
			err = server.ServeTLS(lis, "", "")
		} else {
			err = server.Serve(lis)
		}

		logger.Info("server stopped. an error is always returned", "error", err)
	}()

	go func() {
		signal.Notify(chSignalTerm, os.Interrupt)
		<-chSignalTerm
		if server != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			logger.Info("shutting down server. wait for at most 10 seconds")

			err := server.Shutdown(ctx)
			logger.Info("server closed. an error may be returned", "error", err)
		}
		chExit <- 0
		close(chExit)
	}()

	select {
	case <-chServerStopped:
		logger.Info("server stopped. exiting")
		os.Exit(0)
	case code := <-chExit:
		logger.Info("exit signal received. exiting")
		os.Exit(code)
	}
}

func createTLSConfig(opts options) (*tls.Config, error) {
	if opts.TLSOption != "tls" {
		return nil, nil
	}

	cert, err := tls.LoadX509KeyPair(
		path.Join(os.Getenv("PWD"), "certs/cert.pem"),
		path.Join(os.Getenv("PWD"), "certs/key.pem"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load TLS certificate and key: %w", err)
	}

	return &tls.Config{
		MinVersion:   tls.VersionTLS13,
		Certificates: []tls.Certificate{cert},
	}, nil
}
