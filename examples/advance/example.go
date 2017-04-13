package main

import (
	"fmt"
	"github.com/domac/go-pod/gopod"
	"os"
	"runtime/pprof"
	"time"
)

//继承 PodJob 接口
type AdvanceJob struct {
	idx  int
	msg  string
	data []string
}

func NewSimpleJob(idx int, msg string) *AdvanceJob {
	sj := &AdvanceJob{
		idx: idx,
		msg: msg}

	//模拟分配大块对象内存
	length := 100

	for i := 0; i < length; i++ {
		sj.data = append(sj.data, msg)
	}

	return sj
}

//作业开启
func (job *AdvanceJob) Start() {
	fmt.Printf("[START] %d = %s | size : %d \n", job.idx, job.msg, len(job.data))
}

//作业资源释放
func (job *AdvanceJob) Finish() {
	job.msg = ""
	job.data = job.data[:0]
	fmt.Printf("[FINISH] %d = %s | size : %d \n", job.idx, job.msg, len(job.data))
}

func main() {
	ticker := time.Tick(1 * time.Second)
	exchan := make(chan int)

	advancePod := gopod.NewPod(50) //协程池大小为10
	advancePod.Start()             //开启Pod调度

	go func() {
		for {
			select {
			case <-ticker:
				//每秒模拟派分200个作业
				go func() {
					for i := 0; i < 500; i++ {
						msg := fmt.Sprintf("msg-%d-hello", i)
						jb := NewSimpleJob(i+1, msg)
						advancePod.AddJob(jb)
					}

					p := pprof.Lookup("goroutine")
					fmt.Printf("============> 当前全局goroutinue数量: %d \n", p.Count())
				}()
			case <-exchan:
				advancePod.Stop()
				os.Exit(2)
			}
		}
	}()

	time.Sleep(1 * time.Minute)
	exchan <- 1
}
