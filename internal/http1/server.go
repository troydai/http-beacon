package http1

import (
	"context"
	"net"
	"net/http"

	"github.com/troydai/http-beacon/internal/server"
)

var _ server.ServerLike = (*Server)(nil)

type Server struct {
	httpServer *http.Server
}

func (h *Server) Start(lis net.Listener) error {
	return h.httpServer.Serve(lis)
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
