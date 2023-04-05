package tcp

import (
	"context"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/LynchQ/my-go-redis/interface/tcp"
	"github.com/LynchQ/my-go-redis/lib/logger"
)

type Config struct {
	Address string `cfg:"address"`
}

// ListenAndServeWithSignal 监听并处理请求，带有信号处理
func ListenAndServeWithSignal(cfg *Config, handler tcp.Handler) error {
	closeChan := make(chan struct{})
	sigChan := make(chan os.Signal)
	// 当系统接收到SIGHUP, SIGQUIT, SIGTERM, SIGINT信号时，会向sigChan发送消息
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)

	// 当sigChan接收到消息时，给closeChan发送消息，关闭监听
	go func() {
		sig := <-sigChan
		switch sig {
		case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			closeChan <- struct{}{}
		}
	}()

	listener, err := net.Listen("tcp", cfg.Address)
	if err != nil {
		return err
	}
	logger.Info("start listen on " + cfg.Address)
	ListenAndServer(listener, handler, closeChan)

	return nil
}

// ListenAndServer 监听并处理请求
func ListenAndServer(listener net.Listener, handler tcp.Handler, closeChan <-chan struct{}) error {
	// 当客户端关闭时，关闭监听
	go func() {
		<-closeChan // 如果接收到closeChan的消息，则关闭监听
		logger.Info("shutting down the server")
		_ = listener.Close()
		_ = handler.Close()
	}()

	defer func() {
		_ = listener.Close()
		_ = handler.Close()
	}()

	ctx := context.Background()
	// 2. 当循环braek时，处理好已经接收的连接
	var waitGroup sync.WaitGroup
	// 1. 循环接收连接
	for {
		conn, err := listener.Accept()
		if err != nil {
			break
		}
		logger.Info("accept a new connection")
		// 3. 当循环braek时，处理好已经接收的连接
		waitGroup.Add(1)
		// 新建协程，一个协程一个连接
		go func() {
			// 4. 当循环braek时，处理好已经接收的连接
			defer waitGroup.Done()
			handler.Handle(ctx, conn)
		}()
	}
	waitGroup.Wait()
	return nil
}
