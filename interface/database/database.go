package database

import "github.com/LynchQ/my-go-redis/interface/resp"

// CmdLine是 [][]byte 的别名，表示命令行
type CmdLine = [][]byte

// Database is the interface for database
type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply
	Close()
	AfterClientClose(c resp.Connection)
}

// DataEntity存储绑定到键的数据，包括字符串、列表、哈希、集合等
type DataEntity struct {
	Data interface{}
}
