package main

import (
	"context"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/troydai/http-beacon/internal/http1"
	"github.com/troydai/http-beacon/internal/http2"
	"github.com/troydai/http-beacon/internal/server"
)

type options struct {
	Protocol string `env:"PROTO_OPTION" envDefault:"http2" enums:"http1,http2"`
}

func getHandler(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("request received", "method", r.Method, "path", r.URL.Path)
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Server-Name", "troydai/http-beacon")
		w.WriteHeader(http.StatusOK)
	})
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

	var server server.ServerLike
	switch opts.Protocol {
	case "http1":
		server = http1.NewServer(getHandler(logger))
	case "http2":
		server = http2.NewServer(getHandler(logger))
	default:
		logger.Error("unsupported protocol", "protocol", opts.Protocol)
		os.Exit(1)
	}

	chServerStopped := make(chan struct{})
	chSignalTerm := make(chan os.Signal, 1)
	chExit := make(chan int)

	go func() {
		defer close(chServerStopped)
		logger.Info("server started")
		err := server.Start(lis)
		logger.Info("server stopped. an error is always returned", "error", err)
	}()

	go func() {
		signal.Notify(chSignalTerm, os.Interrupt)
		<-chSignalTerm
		if server != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			logger.Info("shutting down server. wait for at most 10 seconds")

			err := server.Stop(ctx)
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
