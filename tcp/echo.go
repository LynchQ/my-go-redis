package tcp

/**
 * 用于测试服务器是否正常工作的echo服务器
 */

import (
	"bufio"
	"context"
	"io"
	"net"
	"sync"
	"time"

	"github.com/LynchQ/my-go-redis/lib/logger"

	"github.com/LynchQ/my-go-redis/lib/sync/wait"

	"github.com/LynchQ/my-go-redis/lib/sync/atomic"
)

// EchoHandler echos接收到客户的线路，用于测试
type EchoHandler struct {
	activeConn sync.Map       // 活跃的连接
	closing    atomic.Boolean // 是否关闭
}

// MakeEchoHandler 创建EchoHandler
func MakeHandler() *EchoHandler {
	return &EchoHandler{}
}

// MakeEchoHandler 创建EchoHandler 用于测试
type EchoClient struct {
	Conn    net.Conn  // 连接
	Waiting wait.Wait // 等待
}

// Close 关闭连接
func (c *EchoClient) Close() error {
	c.Waiting.WaitWithTimeout(10 * time.Second) // 等待10秒
	c.Conn.Close()
	return nil
}

// Handle echos接收到客户的线路
func (h *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	// 如果正在关闭，拒绝新的连接
	if h.closing.Get() {
		_ = conn.Close()
	}
	// 新建客户端
	client := &EchoClient{
		Conn: conn,
	}
	// 存储客户端
	h.activeConn.Store(client, struct{}{})

	// 读取客户端的数据
	reader := bufio.NewReader(conn)
	// 循环读取
	for {
		// 可能发生的错误：客户端EOF，客户端超时，服务器提前关闭
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				// 客户端关闭连接
				h.activeConn.Delete(client)
				return
			}
			// 其他错误
			logger.Warn("read error: %s", err)
			h.activeConn.Delete(client)
			return
		}
		client.Waiting.Add(1) // 等待
		b := []byte(msg)      // 转换为字节
		_, _ = conn.Write(b)  // 写回客户端
		client.Waiting.Done() // 完成
	}
}

// Close 关闭服务器
func (h *EchoHandler) Close() error {
	logger.Info("handler shutting down...")
	// 设置关闭标志
	h.closing.Set(true)
	// 关闭所有连接
	h.activeConn.Range(func(key interface{}, val interface{}) bool {
		client := key.(*EchoClient)
		_ = client.Close()
		return true
	})
	return nil
}
