package resp

// Reply 是redis序列化协议消息的接口
type Reply interface {
	ToBytes() []byte
}

/* ---- Error Reply ---- */
type ErrorReply interface {
	Error() string
	ToBytes() []byte
}
