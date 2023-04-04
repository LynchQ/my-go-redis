package atomic

import "sync/atomic"

// Boolean 是一个布尔值，它的所有操作都是原子的
type Boolean uint32

// Get 以原子方式读取值
func (b *Boolean) Get() bool {
	return atomic.LoadUint32((*uint32)(b)) != 0
}

// Set 以原子方式写入值
func (b *Boolean) Set(v bool) {
	// 将b的值设置为1 or 0
	if v {
		atomic.StoreUint32((*uint32)(b), 1)
	} else {
		atomic.StoreUint32((*uint32)(b), 0)
	}
}
