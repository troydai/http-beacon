package h2c

import (
	"context"
	"net"
	"net/http"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

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
	s := &http.Server{
		Handler: h2c.NewHandler(handler, &http2.Server{}),
	}

	return &Server{httpServer: s}
}
