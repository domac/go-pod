package gopod

import (
	"sync"
	"time"
)

type MF func(task Task)

//任务分发器
type Dispatcher struct {
	maxExecutors    int
	queueBufferSize int
	wg              *sync.WaitGroup
	wait            bool
	taskQueue       chan Task
	taskPool        chan chan Task
	executors       []Executor
	quit            chan bool
	limiter         <-chan time.Time
	openmq          bool
	messagePipeline chan Task //消息管道
	mpwg            *sync.WaitGroup
	messageFunc     MF
	priority        uint64 //优先执行数

	running bool
}

//创建分发器
func NewDispatcher(maxExecutors, queueBufferSize int) *Dispatcher {
	dispatcher := &Dispatcher{
		maxExecutors:    maxExecutors,
		queueBufferSize: queueBufferSize,
		taskPool:        make(chan chan Task, maxExecutors),
		wait:            false,
		quit:            make(chan bool),
		executors:       make([]Executor, maxExecutors),
		openmq:          false,
		running:         false,
	}

	if queueBufferSize != 0 {
		dispatcher.taskQueue = make(chan Task, queueBufferSize)
	} else {
		dispatcher.taskQueue = make(chan Task)
	}
	return dispatcher
}

//创建带等待的分发器
func NewDispatcherWithWait(maxExecutors, queueBufferSize int, wg *sync.WaitGroup) *Dispatcher {
	dispatcher := NewDispatcher(maxExecutors, queueBufferSize)
	dispatcher.wg = wg
	dispatcher.wait = true
	return dispatcher
}

func NewDispatcherWithMQ(maxExecutors, queueBufferSize int, wg *sync.WaitGroup, mpwg *sync.WaitGroup) *Dispatcher {
	dispatcher := NewDispatcherWithWait(maxExecutors, queueBufferSize, wg)

	dispatcher.SetMF(func(task Task) {
		//println("no data sent")
	})

	dispatcher.openmq = true
	dispatcher.mpwg = mpwg
	dispatcher.messagePipeline = make(chan Task, maxExecutors) //初始化消息管道

	return dispatcher
}

func (dispatcher *Dispatcher) SubmitTask(task Task) {
	if dispatcher.wait {
		dispatcher.wg.Add(1)

		if dispatcher.openmq {
			dispatcher.mpwg.Add(1)
		}

	}
	dispatcher.taskQueue <- task
}

//设置优先执行数
func (dispatcher *Dispatcher) SetPriority(num int) {
	if num < 0 {
		num = 0
	}
	dispatcher.priority = uint64(num)
}

func (dispatcher *Dispatcher) Run() {

	dispatcher.running = true

	for i := 0; i < dispatcher.maxExecutors; i++ {
		if dispatcher.wait {
			if !dispatcher.openmq {
				dispatcher.executors[i] = NewExecutorWithWait(dispatcher.taskPool, dispatcher.wg)
			} else {
				dispatcher.executors[i] = NewExecutorWithMQ(dispatcher.taskPool, dispatcher.messagePipeline, dispatcher.wg)
			}
		} else {
			dispatcher.executors[i] = NewExecutor(dispatcher.taskPool)
		}
		//开启执行
		dispatcher.executors[i].Start()
	}

	//开启任务分发
	go dispatcher.dispatch()

	//消息上报
	if dispatcher.openmq {
		go dispatcher.report()
	}

}

func (dispatcher *Dispatcher) RunWithLimiter(limiterGap time.Duration) {
	dispatcher.limiter = time.Tick(limiterGap)
	dispatcher.Run()
}

//任务分发处理
func (dispatcher *Dispatcher) dispatch() {
	defer dispatcher.shutdown()
	index := uint64(0)
	for {
		if dispatcher.priority >= 0 {
			index++
		}
		select {
		case task := <-dispatcher.taskQueue:

			executorTaskChan := <-dispatcher.taskPool

			if dispatcher.limiter != nil && index > dispatcher.priority {
				index = dispatcher.priority //防止index越界
				<-dispatcher.limiter
			}

			executorTaskChan <- task

		case <-dispatcher.quit:
			return
		}
	}
}

//消息上报
func (dispatcher *Dispatcher) report() {
	for {
		select {
		case messageTask := <-dispatcher.messagePipeline:
			f := dispatcher.messageFunc

			//并行处理数据上报
			go func() {
				f(messageTask)
				dispatcher.mpwg.Done()
			}()

		case <-dispatcher.quit:
			return
		}
	}
}

func (dispatcher *Dispatcher) shutdown() {
	dispatcher.running = false
	for _, e := range dispatcher.executors {
		for !e.Stop() { //一直处理停止
		}
	}
	close(dispatcher.taskPool)
	close(dispatcher.taskQueue)
	close(dispatcher.messagePipeline)
}

//停止开关
func (dispatcher *Dispatcher) Stop() {
	dispatcher.running = false
	dispatcher.quit <- true
}

//设置消息处理方法
func (dispatcher *Dispatcher) SetMF(mf MF) {
	dispatcher.messageFunc = mf
}

func (dispatcher *Dispatcher) IsRunning() bool {
	return dispatcher.running
}
