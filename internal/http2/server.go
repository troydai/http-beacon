package http2

import (
	"context"
	"net"
	"net/http"
	"os"
	"path"

	"github.com/troydai/http-beacon/internal/server"
)

var _ server.ServerLike = (*Server)(nil)

type Server struct {
	httpServer *http.Server
}

// Server implements server.ServerLike.
func (h *Server) Start(lis net.Listener) error {
	return h.httpServer.ServeTLS(
		lis,
		path.Join(os.Getenv("PWD"), "certs/cert.pem"),
		path.Join(os.Getenv("PWD"), "certs/key.pem"),
	)
}

func (h *Server) Stop(ctx context.Context) error {
	return h.httpServer.Shutdown(ctx)
}

func NewServer(handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Handler: handler,
		},
	}
}
