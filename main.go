package main

import (
	"context"
	"crypto/tls"
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
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalf("error create logger: %v", err)
	}

	lc := &net.ListenConfig{KeepAlive: -1}

	lis, err := lc.Listen(context.Background(), "tcp", "127.0.0.1:8443")
	if err != nil {
		logger.Fatal("error start TCP listener", zap.Error(err))
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

func customizeConnContext(logger *zap.Logger) func(context.Context, net.Conn) context.Context {
	if logger == nil {
		logger = zap.NewNop()
	}

	return func(ctx context.Context, c net.Conn) context.Context {
		logger.Info("connection context called.")
		if tls, ok := c.(*tls.Conn); ok {
			logger.Info("tls connection established.", zap.String("local", tls.LocalAddr().String()), zap.String("remote", tls.RemoteAddr().String()))

			if tcp, ok := tls.NetConn().(*net.TCPConn); ok {
				logger.Info("tcp connection type casted. set keep alive to false.")
				if err := tcp.SetKeepAlive(false); err != nil {
					logger.Error("error set keep alive to false.", zap.Error(err))
				}
			}
		}

		return ctx
	}
}
