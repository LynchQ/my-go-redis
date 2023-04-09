package parser

import (
	"bufio"
	"errors"
	"io"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/LynchQ/my-go-redis/interface/resp"
	"github.com/LynchQ/my-go-redis/lib/logger"
	"github.com/LynchQ/my-go-redis/resp/reply"
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

// 解析器的入口，异步解析
func ParseStream(reader io.Reader) <-chan *Payload {
	// 1. 创建一个 channel
	ch := make(chan *Payload)
	// 2. 启动一个 goroutine，调用 parse0 函数
	go parse0(reader, ch)
	// 3. 返回 channel
	return ch
}

// 解析器核心
func parse0(reader io.Reader, ch chan<- *Payload) {
	defer func() {
		// 如果解析过程中出现 panic，就会在这里捕获
		// recover() 会返回 panic 的值
		if err := recover(); err != nil {
			logger.Error(string(debug.Stack()))
		}
	}()

	// 创建一个 bufio.Reader 对象 用于读取数据
	bufReader := bufio.NewReader(reader)
	var state readState // 解析器的状态
	var err error       //	错误
	var msg []byte      // 读取到的数据

	for {
		// 1. 读取一个字节
		var ioErr bool // 是否是 io 错误
		msg, ioErr, err = readLine(bufReader, &state)
		if err != nil {
			if ioErr { // 如果是 io 错误，就关闭 channel，结束解析
				ch <- &Payload{
					Err: err,
				}
				close(ch)
				return
			}
			// 如果是协议错误，重置状态，继续解析
			ch <- &Payload{
				Err: err,
			}
			state = readState{}
			continue
		}

		// 2. 根据字节类型，调用不同的解析函数
		if !state.redingMultiLine { // 没有在读取多行, 也许是还没有开始解析
			if msg[0] == '*' {
				// 读取到 * 开头的，说明是多行
				// 解析多行
				err = parseMultiBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{
						Err: errors.New("protocol error: " + string(msg)),
					}
					state = readState{}
					continue
				}
				// 如果参数个数是 0，就直接把解析结果放到 channel 中
				if state.expectedArgsCount == 0 {
					ch <- &Payload{
						Data: &reply.EmptyMultiBulkReply{},
					}
					state = readState{}
					continue
				}
			} else if msg[0] == '$' {
				// 读取到 $ 开头的，说明是多行的一部分
				err = parseBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{
						Err: errors.New("protocol error: " + string(msg)),
					}
					state = readState{}
					continue
				}
				if state.bulkLen == -1 {
					ch <- &Payload{
						Data: &reply.NullBulkReply{},
					}
					state = readState{}
					continue
				}
			} else {
				// 读取到 + 开头的，说明是单行
				// 解析单行
				result, err := parseSingleLineReply(msg)

				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{}
				continue
			}
		} else {
			// 3. 如果是多行，就继续读取
			err = readBody(msg, &state)
			if err != nil {
				ch <- &Payload{
					Err: errors.New("protocol error: " + string(msg)),
				}
				state = readState{}
				continue
			}
			// 4. 如果读取完成，就把解析结果放到 channel 中
			if state.finished() {
				var result resp.Reply
				if state.msgType == '*' {
					result = &reply.MultiBulkReply{Args: state.args}
				} else if state.msgType == '$' {
					result = &reply.BulkReply{Arg: state.args[0]}
				}
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{} // 重置状态
			}
		}
	}
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

// parseSingleLineReply
func parseSingleLineReply(msg []byte) (resp.Reply, error) {
	str := strings.TrimSuffix(string(msg), "\r\n") // 去掉 \r\n

	var result resp.Reply
	// 根据第一个字符，判断是什么类型的回复
	switch msg[0] {
	case '+': // 状态回复
		result = reply.MakeStatusReply(str[1:])
	case '-': // 错误回复
		result = reply.MakeErrReply(str[1:])

	case ':': // 整数回复
		val, err := strconv.ParseInt(str[1:], 10, 64) // strconv.ParseInt 用来解析整数
		if err != nil {
			return nil, errors.New("protocol error: " + string(msg))
		}
		result = reply.MakeIntReply(val)
	}
	return result, nil
}

// 读取多批量回复或批量回复的非第一行
func readBody(msg []byte, state *readState) error {
	line := msg[0 : len(msg)-2] // 去掉 \r\n
	var err error

	if line[0] == '$' { // 批量回复
		// bulk reply 批量回复
		state.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 64) // 10进制 64位
		if err != nil {
			return errors.New("protocol error: " + string(msg)) // 协议错误
		}
		if state.bulkLen < 0 { // 如果长度是 -1，就是空值
			state.args = append(state.args, []byte{})
			state.bulkLen = 0
		} else {
			state.args = append(state.args, line[1:])
		}
	}
	return nil
}
