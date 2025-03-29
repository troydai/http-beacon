package server

import (
	"context"
	"net"
)

type ServerLike interface {
	Start(lis net.Listener) error
	Stop(ctx context.Context) error
}
