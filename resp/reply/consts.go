package reply

// PongReply is +PONG
type PongReply struct{}

var pongBytes = []byte("+PONG\r\n")

// ToBytes marshal redis.Reply
func (r PongReply) ToBytes() []byte {
	return pongBytes
}

func MakePongReply() *PongReply {
	return &PongReply{}
}

// OkReply is +OK
type OkReply struct{}

var okBytes = []byte("+OK\r\n")

// ToBytes marshal redis.Reply
func (r OkReply) ToBytes() []byte {
	return okBytes
}

func MakeOkReply() *OkReply {
	return &OkReply{}
}

// NullBulkReply 空字符串
type NullBulkReply struct{}

var nullBulkBytes = []byte("$-1\r\n")

func (r NullBulkReply) ToBytes() []byte {
	return nullBulkBytes
}

func MakeNullBulkReply() *NullBulkReply {
	return &NullBulkReply{}
}

// EmptyMultiBulkReply 空数组
type EmptyMultiBulkReply struct{}

var emptymultibulkbytes = []byte("*0\r\n")

func (r EmptyMultiBulkReply) ToBytes() []byte {
	return emptymultibulkbytes
}

func MakeEmptyMultiBulkReply() *EmptyMultiBulkReply {
	return &EmptyMultiBulkReply{}
}

// NoReply 什么都不返回
type NoReply struct{}

var noBytes = []byte("")

func (r NoReply) ToBytes() []byte {
	return noBytes
}

func MakeNoReply() *NoReply {
	return &NoReply{}
}
