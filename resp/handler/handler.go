package handler

import (
	"context"
	"io"
	"net"
	"strings"
	"sync"

	"github.com/LynchQ/my-go-redis/database"
	databaseface "github.com/LynchQ/my-go-redis/interface/database"
	"github.com/LynchQ/my-go-redis/lib/logger"
	"github.com/LynchQ/my-go-redis/lib/sync/atomic"
	"github.com/LynchQ/my-go-redis/resp/connection"
	"github.com/LynchQ/my-go-redis/resp/parser"
	"github.com/LynchQ/my-go-redis/resp/reply"
)

/*
 * tcp.RespHandler实现redis协议
 */
// RespHandler is the handler for RESP protocol
var (
	unknownErrReplyBytes = []byte("-ERR unknown\r\n")
)

type RespHandler struct {
	activeConn sync.Map // *client -> placeholder
	db         databaseface.Database
	closing    atomic.Boolean // 拒绝新客户端和新请求
}

// MakeHandler创建RespHandler实例
func MakeHandler() *RespHandler {
	// var db databaseface.Database
	db := database.NewEchoDatabase()
	return &RespHandler{
		db: db,
	}
}

func (h *RespHandler) closeClient(client *connection.Connection) {
	_ = client.Close()            // 关闭客户端
	h.db.AfterClientClose(client) // 关闭数据库
	h.activeConn.Delete(client)   // 删除客户端
}

// Handle接收并执行redis命令
func (h *RespHandler) Handle(ctx context.Context, conn net.Conn) {
	if h.closing.Get() {
		// 关闭处理程序拒绝新连接
		_ = conn.Close()
	}

	// 创建客户端
	client := connection.NewConn(conn)
	h.activeConn.Store(client, 1) // 存储客户端

	// 解析流
	ch := parser.ParseStream(conn)
	for payload := range ch {
		// 解析错误
		if payload.Err != nil {
			if payload.Err == io.EOF || // 客户端主动关闭
				payload.Err == io.ErrUnexpectedEOF || // 客户端关闭 io.ErrUnexpectedEOF 的错误信息是 "unexpected EOF"
				// strings.Contains 判断字符串是否包含某个字符串
				strings.Contains(payload.Err.Error(), "use of closed network connection") {
				// 客户端关闭
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			// 协议错误
			errReply := reply.MakeErrReply(payload.Err.Error())
			err := client.Write(errReply.ToBytes())
			if err != nil {
				h.closeClient(client)
				logger.Info("connection closed: " + client.RemoteAddr().String())
				return
			}
			continue
		}
		// 解析数据
		if payload.Data == nil {
			logger.Error("empty payload")
			continue
		}
		reply, ok := payload.Data.(*reply.MultiBulkReply)
		// q: 为什么要转换成MultiBulkReply
		// a: 因为redis的命令都是以数组的形式传输的
		if !ok {
			logger.Error("request muti bulk reply")
			continue
		}
		// 执行命令 Exec
		result := h.db.Exec(client, reply.Args)
		if result != nil {
			_ = client.Write(result.ToBytes())
		} else {
			_ = client.Write(unknownErrReplyBytes)
		}
	}
}

// Close关闭处理程序
func (h *RespHandler) Close() error {
	logger.Info("handler shutting down...")
	// 拒绝新连接
	h.closing.Set(true)
	// TODO concurrent wait

	h.activeConn.Range(func(key, value interface{}) bool {
		client := key.(*connection.Connection)
		_ = client.Close()
		return true
	})
	h.db.Close()
	return nil
}
