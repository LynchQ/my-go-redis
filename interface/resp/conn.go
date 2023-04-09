package resp

// Connection 表示与redis客户端的连接
type Connection interface {
	Write([]byte) error // 写入数据
	GetDBIndex() int    // 用于多数据库
	SelectDB(int)       // 用于切换数据库
}
