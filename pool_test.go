package task_pool

import (
	"fmt"
	"testing"
	"time"
)

func TestPool(t *testing.T) {

	pool := NewPool(100, 2, 0.75, 0.25, 100)

	for i := 0; i < 1000; i++ {
		var sleepTime time.Duration
		if i < 100 {
			// 大洪峰
			sleepTime = time.Duration(i) * time.Millisecond
		} else if i > 800 {
			// 这里大于800时主要模拟洪峰已过 测试缩容
			sleepTime = 300 * time.Millisecond
		} else {
			// 小洪峰
			sleepTime = 100 * time.Millisecond
		}
		time.Sleep(sleepTime)

		b := i
		pool.AddTask(func() {
			time.Sleep(time.Duration(b) * time.Millisecond)
			fmt.Println("execute task", b, "done.", "tasks:", len(pool.taskChannel))
		})
	}

	for {
		poolTaskSize := len(pool.taskChannel)
		fmt.Println("task size:", poolTaskSize)
		if poolTaskSize == 0 {
			return
		}
		time.Sleep(time.Minute)
	}

}
