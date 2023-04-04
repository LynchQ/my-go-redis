package tcp

import (
	"context"
	"net"

	"github.com/LynchQ/my-go-redis/interface/tcp"
	"github.com/LynchQ/my-go-redis/lib/logger"
)

type Config struct {
	Address string `cfg:"address"`
}

// ListenAndServeWithSignal 监听并处理请求，带有信号处理
func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {
	closeChan := make(chan struct{})
	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	logger.Info("start listen")
	ListenAndServer(listener, handler, closeChan)

	return nil
}

// ListenAndServer 监听并处理请求
func ListenAndServer(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) error {
	ctx := context.Background()
	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		logger.Info("accept a new connection")
		// 新建协程，一个协程一个连接
		go func() {
			handler.Handle(ctx, conn)
		}()
	}
	return nil
}
