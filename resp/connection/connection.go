package connection

import (
	"net"
	"sync"
	"time"

	"github.com/LynchQ/my-go-redis/lib/sync/wait"
)

// Connection 表示使用redis-cli的连接
type Connection struct {
	conn         net.Conn   // 与客户端的连接
	waitingReply wait.Wait  // 等待回复完成
	mu           sync.Mutex // 处理发送响应时的锁
	selectedDB   int        // 选择的数据库
}

// NewConn 创建一个新的连接 接收一个net.Conn 作为参数 返回一个指向Connection的指针
func NewConn(conn net.Conn) *Connection {
	return &Connection{
		conn: conn,
	}
}

// RemoteAddr 返回远程网络地址 返回一个net.Addr
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// Close 断开与客户端的连接
func (c *Connection) Close() error {
	// 等待10秒
	c.waitingReply.WaitWithTimeout(10 * time.Second)
	_ = c.conn.Close()
	return nil
}

// Write 通过tcp连接向客户端写入发送响应
func (c *Connection) Write(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	c.mu.Lock()           // 1.加锁
	c.waitingReply.Add(1) // 2.等待回复完成
	defer func() {
		c.waitingReply.Done() // 3.等待回复完成
		c.mu.Unlock()         // 4.解锁
	}()

	_, err := c.conn.Write(b)
	return err
}

// GetDBIndex 返回选择的数据库
func (c *Connection) GetDBIndex() int {
	return c.selectedDB
}

// SelectDB 选择一个数据库
func (c *Connection) SelectDB(dbNum int) {
	c.selectedDB = dbNum
}
