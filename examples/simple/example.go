package main

import (
	"fmt"
	"github.com/domac/go-pod/gopod"
	"time"
)

//继承 PodJob 接口
type SimpleJob struct {
	idx int
	msg string
}

func NewSimpleJob(idx int, msg string) *SimpleJob {
	sj := &SimpleJob{
		idx: idx,
		msg: msg}
	return sj
}

//作业开启
func (job *SimpleJob) Start() {
	fmt.Printf("[START] %d = %s \n", job.idx, job.msg)
	time.Sleep(1000 * time.Millisecond)
}

//作业释放
func (job *SimpleJob) Finish() {
	job.msg = ""
	fmt.Printf("[FINISH] %d = %s \n", job.idx, job.msg)
}

func main() {
	simplePod := gopod.NewPod(10) //协程池大小为10
	simplePod.Start()             //开启Pod调度
	go func() {
		//模拟分配作业
		for i := 0; i < 100; i++ {
			msg := fmt.Sprintf("msg-%d-hello", i)
			jb := NewSimpleJob(i+1, msg)
			simplePod.AddJob(jb)
		}
	}()
	time.Sleep(500 * time.Second)
}
