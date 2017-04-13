package gopod

import (
	"errors"
	"sync"
	"time"
)

//作业接口
type PodJob interface {
	Start()
	Finish()
}

type Pod struct {
	dispatcher *Dispatcher
	gomaxsize  int
	interval   time.Duration
}

func NewPod(gomaxsize int) *Pod {
	interval := time.Millisecond * 2 //默认调度间隔为2毫秒
	d := NewDispatcher(gomaxsize, gomaxsize)
	pod := &Pod{
		dispatcher: d,
		gomaxsize:  gomaxsize,
		interval:   interval,
	}
	return pod
}

func NewPodWithInterval(gomaxsize int, interval time.Duration) *Pod {
	d := NewDispatcher(gomaxsize, gomaxsize)
	pod := &Pod{
		dispatcher: d,
		gomaxsize:  gomaxsize,
		interval:   interval,
	}
	return pod
}

//启动Pod
func (self *Pod) Start() error {

	if self.dispatcher == nil {
		return errors.New("pod dispatcher is not init")
	}

	if self.dispatcher.IsRunning() == true {
		return errors.New("pod dispatcher was already start")
	}

	self.dispatcher.RunWithLimiter(self.interval)
	return nil
}

func (self *Pod) Stop() error {
	//关闭本轮调度

	if self.dispatcher.IsRunning() == false {
		return errors.New("pod dispatcher was already stop")
	}

	self.dispatcher.Stop()
	return nil
}

func (self *Pod) AddJob(pjob PodJob) {
	task := CreateTask(pjob, "Start", "Finish")
	self.dispatcher.SubmitTask(task)
}

func (self *Pod) AddJobs(pjobs []PodJob) {
	for _, pj := range pjobs {
		self.AddJob(pj)
	}
}

//同步包装器
type WaitGroupWrapper struct {
	sync.WaitGroup
}

//simple function wrapper
func (w *WaitGroupWrapper) Wrap(sf func()) {
	w.Add(1)
	go func() {
		sf()
		w.Done()
	}()
}
