package reply

import (
	"bytes"
	"strconv"

	"github.com/LynchQ/my-go-redis/interface/resp"
)

var (
	nullBulkReplyBytes = []byte("$-1")

	// CRLF是redis序列化协议的行分隔符
	CRLF = "\r\n"
)

/* ---- Bulk Reply ---- */
// BulkReply存储二进制安全字符串

type BulkReply struct {
	Arg []byte // "LynchQ" "$6\r\nLynchQ\r\n"
}

// MakeBulkReply创建BulkReply
func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{
		Arg: arg,
	}
}

func (r *BulkReply) ToBytes() []byte {
	if len(r.Arg) == 0 {
		return nullBulkReplyBytes
	}
	// strconv.Itoa() 函数用于将整型转换为字符串
	return []byte("$" + strconv.Itoa(len(r.Arg)) + CRLF + string(r.Arg) + CRLF)
}

/* ---- Multi Bulk Reply ---- */
// MultiBulkReply存储字符串列表

type MultiBulkReply struct {
	Args [][]byte // "SET name LynchQ" "*3\r\n$3\r\nSET\r\n$4\r\nname\r\n$6\r\nLynchQ\r\n"
}

func (r *MultiBulkReply) ToBytes() []byte {
	argLen := len(r.Args)
	// 利用bytes.Buffer实现字符串拼接 比+号拼接效率高
	// bytes.Buffer是一个实现了读写方法的可变大小的字节缓冲
	var buf bytes.Buffer
	// WriteString将字符串写入缓冲，返回写入的字节数和可能遇到的任何错误
	buf.WriteString("*" + strconv.Itoa(argLen) + CRLF)

	for _, arg := range r.Args {
		if arg == nil {
			buf.WriteString("$-1" + CRLF)
		} else {
			buf.WriteString("$" + strconv.Itoa(len(arg)) + CRLF + string(arg) + CRLF)
		}
	}
	// Bytes返回缓冲中的数据
	return buf.Bytes()
}

// MakeMultiBulkReply创建MultiBulkReply
func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{
		Args: args,
	}
}

/* ---- Status Reply ---- */
// StatusReply存储简单状态字符串
type StatusReply struct {
	Status string
}

func (r *StatusReply) ToBytes() []byte {
	return []byte("+" + r.Status + CRLF)
}

// MakeStatusReply创建StatusReply
func MakeStatusReply(status string) *StatusReply {
	return &StatusReply{
		Status: status,
	}
}

/* ---- Integer Reply ---- */
// IntegerReply存储整数
type IntReply struct {
	Code int64
}

func (r *IntReply) ToBytes() []byte {
	// strconv.FormatInt() 函数用于将整型转换为字符串
	return []byte(":" + strconv.FormatInt(r.Code, 10) + CRLF)
}

// MakeIntReply创建IntReply
func MakeIntReply(code int64) *IntReply {
	return &IntReply{
		Code: code,
	}
}

// StandardReply 标准错误回复
type StandardErrReply struct {
	Status string
}

func (r *StandardErrReply) ToBytes() []byte {
	return []byte("-" + r.Status + CRLF)
}

func (r *StandardErrReply) Error() string {
	return r.Status
}

// MakeErrReply创建StandardErrReply
func MakeErrReply(status string) *StandardErrReply {
	return &StandardErrReply{
		Status: status,
	}
}

// IsErrReply 返回是否是错误回复
func IsErrReply(reply resp.Reply) bool {
	// reply.ToBytes() 返回[]byte类型 ToBytes()[0] 返回第一个字节
	return reply.ToBytes()[0] == '-'
}
