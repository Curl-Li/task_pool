package task_pool

import (
	"fmt"
	"sync"
)

// Pool 协程池
type Pool struct {
	// maxPoolSize 最大协程池数量
	maxPoolSize int
	// stepLength 扩容和缩容跨度
	stepLength int8
	// expansionCoefficient 扩容系数 len(channel)/cap(channel)>expansionCoefficient 时扩容
	expansionCoefficient float32
	// reduceCoefficient 缩容系数 len(channel)/cap(channel)<expansionCoefficient 时缩容
	reduceCoefficient float32
	// taskChannelSize 任务管道长度
	taskChannelSize int

	stopChannels []chan struct{}
	taskChannel  chan func()
	mutex        sync.Mutex
}

func NewPool(maxPoolSize int, stepLength int8, expansionCoefficient float32, reduceCoefficient float32, taskChannelSize int) *Pool {
	p := &Pool{
		maxPoolSize:          maxPoolSize,
		stepLength:           stepLength,
		expansionCoefficient: expansionCoefficient,
		reduceCoefficient:    reduceCoefficient,
		taskChannelSize:      taskChannelSize,
		taskChannel:          make(chan func(), taskChannelSize),
		mutex:                sync.Mutex{},
	}
	p.mutex.Lock()
	p.expand()
	p.mutex.Unlock()
	return p
}

// AddTask 添加任务
func (p *Pool) AddTask(t func()) {
	p.mutex.Lock()
	length := len(p.taskChannel) + 1
	coefficient := float32(float64(length) / float64(p.taskChannelSize))

	fmt.Println("add task, coefficient:",coefficient)

	if coefficient > p.expansionCoefficient && length+int(p.stepLength) <= p.maxPoolSize {
		// 大于扩容系数并且当前长度+步长小于等于最大poolSize 执行扩容
		p.expand()
	} else if coefficient < p.reduceCoefficient && len(p.stopChannels) > int(p.stepLength) {
		// 小于缩容系数并且协程数量大于步长 执行缩容
		p.reduce()
	}
	p.mutex.Unlock()
	p.taskChannel <- t
}

func (p *Pool) expand() {
	fmt.Println("expand called. ")
	// 执行扩容
	// 初始化stepLength个stopChannel 并go出协程 将stopChannels放入管理对象
	stopFlags := make([]chan struct{}, p.stepLength)
	for i := 0; i < int(p.stepLength); i++ {
		stopFlags[i] = make(chan struct{})
		p.doGo(stopFlags[i], len(p.stopChannels)+i)
	}

	p.stopChannels = append(p.stopChannels, stopFlags...)
	fmt.Println("expand done, routine size: ", len(p.stopChannels))
}

func (p *Pool) reduce() {
	fmt.Println("reduce called. ")
	// 执行缩容 从stopChannels最后面选择关闭stepLength个channel
	// 并将其从对象中删除
	endIndex := len(p.stopChannels) - int(p.stepLength) - 1
	for i := len(p.stopChannels) - 1; i > endIndex; i-- {
		close(p.stopChannels[i])
	}
	p.stopChannels = p.stopChannels[:endIndex+1]
	fmt.Println("reduce done, routine size: ", len(p.stopChannels))
}

func (p *Pool) doGo(stopChannel chan struct{}, routineIndex int) {
	go func() {
	LOOP:
		for {
			select {
			case <-stopChannel:
				// stopChannel不阻塞 说明关闭了 协程结束退出循环
				break LOOP
			case t := <-p.taskChannel:
				// 拿到task 执行
				t()
			}
		}
		fmt.Println("routine: ", routineIndex, "has stopped. ")
	}()
}
