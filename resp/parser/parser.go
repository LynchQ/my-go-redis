package parser

import (
	"io"

	"github.com/LynchQ/my-go-redis/interface/resp"
)

// 客户端发送的命令 和 我们回复的消息 都是 resp.Reply 类型的
type Payload struct {
	Data resp.Reply // 客户端发送的命令 或者 我们回复的消息
	Err  error      // 如果有错误，就会在这里
}

// 解析器的状态
type readState struct {
	redingMultiLine   bool     // 是否正在读取多行
	expectedArgsCount int      // 期望的参数个数
	msgType           byte     // 消息类型
	args              [][]byte // 命令参数
	bulkLen           int64    // 长度
}

// 计算解析有没有完成
func (s *readState) finished() bool {
	return s.expectedArgsCount > 0 && len(s.args) == s.expectedArgsCount
}

// 解析器核心
func parse0(reader io.Reader, ch chan<- *Payload) {
	// 1. 读取一个字节
	// 2. 根据字节类型，调用不同的解析函数
	// 3. 解析完成后，把解析结果放到 channel 中
}

// 解析器的入口，异步解析
func ParseStream(reader io.Reader) <-chan *Payload {
	// 1. 创建一个 channel
	ch := make(chan *Payload)
	// 2. 启动一个 goroutine，调用 parse0 函数
	go parse0(reader, ch)
	// 3. 返回 channel
	return ch
}
