package database

import (
	"github.com/LynchQ/my-go-redis/interface/resp"
	"github.com/LynchQ/my-go-redis/lib/logger"
	"github.com/LynchQ/my-go-redis/resp/reply"
)

type EchoDatabase struct {
}

func NewEchoDatabase() *EchoDatabase {
	return &EchoDatabase{}
}

// Exec 执行命令
func (e EchoDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {
	return reply.MakeMultiBulkReply(args)
}

// AfterClientClose 在客户端关闭后调用
func (e EchoDatabase) AfterClientClose(c resp.Connection) {
	logger.Info("EchoDatabase AfterClientClose")
}

// Close 关闭数据库
func (e EchoDatabase) Close() {
	logger.Info("EchoDatabase Close")
}
