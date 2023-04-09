package reply

/* ---- unknowErr Reply ---- */
type UnknownErrReply struct{}

var unknownErrBytes = []byte("-ERR unknown error\r\n")

func (r UnknownErrReply) Error() string {
	return string(unknownErrBytes)
}

func (r UnknownErrReply) ToBytes() []byte {
	return unknownErrBytes
}

/* ---- ArgNumErr Reply ---- */
type ArgNumErrReply struct {
	Cmd string // 命令
}

// 动态回复
func (r *ArgNumErrReply) Error() string {
	return "ERR wrong number of arguments for '" + r.Cmd + "' command"
}

func (r *ArgNumErrReply) ToBytes() []byte {
	return []byte("-ERR wrong number of arguments for '" + r.Cmd + "' command\r\n")
}

// MakeArgNumErrReply 创建错误回复
func MakeArgNumErrReply(cmd string) *ArgNumErrReply {
	return &ArgNumErrReply{
		Cmd: cmd,
	}
}

/* ---- SyntaxErr Reply ---- */
type SyntaxErrReply struct{}

var syntaxErrBytes = []byte("-ERR syntax error\r\n")
var theSyntaxErrReply = &SyntaxErrReply{}

func (r *SyntaxErrReply) Error() string {
	return "ERR syntax error"
}

func (r *SyntaxErrReply) ToBytes() []byte {
	return syntaxErrBytes
}

// MakeSyntaxErrReply 创建语法错误回复
func MakeSyntaxErrReply() *SyntaxErrReply {
	return theSyntaxErrReply
}

/* ---- WrongTypeErr Reply ---- */
type WrongTypeErrReply struct{}

var wrongTypeErrBytes = []byte("-WRONGTYPE Operation against a key holding the wrong kind of value\r\n")

func (r *WrongTypeErrReply) Error() string {
	return "WRONGTYPE Operation against a key holding the wrong kind of value"
}

func (r *WrongTypeErrReply) ToBytes() []byte {
	return wrongTypeErrBytes
}

/* ---- ProtocolErr Reply ---- */
// 表示在分析请求期间遇到意外字节
type ProtocolErrReply struct {
	Msg string
}

func (r *ProtocolErrReply) Error() string {
	return "Protocol error: " + r.Msg
}

func (r *ProtocolErrReply) ToBytes() []byte {
	return []byte("-Protocol error: " + r.Msg + CRLF)
}
