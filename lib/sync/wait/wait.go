package wait

/**
 * 自己实现waitgroup，用于等待所有的goroutine完成 以及 超时
 */
import (
	"sync"
	"time"
)

type Wait struct {
	wg sync.WaitGroup // 用于等待所有的goroutine完成
}

// Add 将delta（可能为负数）添加到WaitGroup计数器中。
func (w *Wait) Add(delta int) {
	w.wg.Add(delta)
}

// Done 将WaitGroup计数器减1
func (w *Wait) Done() {
	w.wg.Done()
}

// Wait 阻塞，直到WaitGroup计数器为0。
func (w *Wait) Wait() {
	w.wg.Wait()
}

// WaitWithTimeout 阻塞，直到WaitGroup计数器为0或超时
func (w *Wait) WaitWithTimeout(timeout time.Duration) bool {
	// 用于通知goroutine完成
	c := make(chan bool, 1)
	// 启动goroutine，等待所有的goroutine完成
	go func() {
		defer close(c)
		w.wg.Wait()
		c <- true
	}()
	// 等待goroutine完成或超时
	select {
	case <-c:
		return false // 正常完成
	case <-time.After(timeout):
		return true // 超时
	}
}
