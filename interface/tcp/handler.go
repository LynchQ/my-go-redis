package tcp

import (
	"context"
	"net"
)

// Handler 处理器接口
type Handler interface {
	// Handle 处理请求
	Handle(ctx context.Context, conn net.Conn)
	// Close 处理关闭
	Close() error
}
