package parser

import (
	"bufio"
	"errors"
	"io"
	"strconv"

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

func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	// eg: *3\r\n$3\r\nSET\r\n$3\r\nkey\r\n$5\r\nvalue\r\n
	// 不能简单的使用 ReadBytes('\r\n')，因为可能会读到一半
	var msg []byte
	var err error

	// 没有预设长度，就按照 \r\n 来读取
	if state.bulkLen == 0 {
		// 1. 读取一行
		msg, err = bufReader.ReadBytes('\n')
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
	} else {
		// 2. 如果是多行，就继续读取
		// 有预设长度，就按照预设长度来读取
		msg = make([]byte, state.bulkLen+2)
		// io.ReadFull 会尝试读取指定长度的数据，如果读取不到，就会返回错误
		// 读取到的数据会放到 msg 中
		_, err = io.ReadFull(bufReader, msg)
		if err != nil {
			return nil, true, err
		}
		if len(msg) == 0 || msg[len(msg)-2] != '\r' || msg[len(msg)-1] != '\n' {
			return nil, false, errors.New("protocol error: " + string(msg))
		}
		state.bulkLen = 0
	}

	return msg, false, nil
}

// parseMultiBulkHeader 用来改变解析器的状态 (解析多行)
func parseMultiBulkHeader(msg []byte, state *readState) error {
	var err error
	var expectedLine uint64 // 期望的行数

	// 1. 解析出期望的行数
	expectedLine, err = strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32) // 10进制 32位
	if err != nil {
		return errors.New("protocol error: " + string(msg)) // 协议错误
	}

	// 2. 根据期望的行数，改变解析器的状态
	if expectedLine == 0 {
		state.expectedArgsCount = 0
		return nil
	} else if expectedLine > 0 {
		// 期望的行数大于 0，就是多行
		state.msgType = msg[0]                       // 消息类型
		state.redingMultiLine = true                 // 是否正在读取多行
		state.expectedArgsCount = int(expectedLine)  // 期望的参数个数
		state.args = make([][]byte, 0, expectedLine) // 命令参数
		return nil
	} else {
		return errors.New("protocol error: " + string(msg)) // 协议错误
	}
}

// parseBulkHeader 用来改变解析器的状态 (解析字符串)
func parseBulkHeader(msg []byte, state *readState) error {
	var err error
	// 1. 解析出长度
	state.bulkLen, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64) // 10进制 64位
	if err != nil {
		return errors.New("protocol error: " + string(msg)) // 协议错误
	}

	// 2. 如果长度是 -1，就是空值
	if state.bulkLen == -1 {
		state.args = append(state.args, nil)
		return nil
	} else if state.bulkLen > 0 {
		state.msgType = msg[0]            // 消息类型
		state.redingMultiLine = true      // 是否正在读取多行
		state.expectedArgsCount = 1       // 期望的参数个数
		state.args = make([][]byte, 0, 1) // 命令参数
		return nil
	} else {
		return errors.New("protocol error: " + string(msg)) // 协议错误
	}
}
