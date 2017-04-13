go-pod
=========

go-pod是一个使用方便的goroutinue池调度工具库, 目的是为了防止日常开发中goroutinue泛滥调用,形成全局goroutinue过多最后导致泄漏的风险.

### 特性:

- 高并发调度,通过goroutinue池,控制全局的作业协程数,防止 go leak
- 支持自定义协程池大小
- 支持自定义作业调度间隔


## 具体实例

本库已为下面的工具提供调度支持

| 项目名称  | 描述 |
| ----- | ----------- |
| [racoon](http://gitlab.tools.vipshop.com/devops-team/racoon) | snmp采集器 |
| [tango](http://gitlab.tools.vipshop.com/devops-team/devops_tools/tree/master/tango) | virgo升级工具 |
| [ptoe](http://gitlab.tools.vipshop.com/vipcloud/app-conf-manage/tree/master/ptoe) | Puppet to Etcd转换工具

## 如何使用

### 获取本库

```
$ cd $GOPATH/src/gitlab/gitlab.tools.vipshop.com/devops-team

$ git clone git@gitlab.tools.vipshop.com:devops-team/go-pod.git
```

### 使用前说明

典型调用方式

```go
pod := gopod.NewPod(10) //协程池大小为10
pod.Start()             //开启Pod调度
pod.AddJob(jb)
```

其中jb为自定义的作业, 需要继承 `PodJob` 接口

```go
type PodJob interface {
	Start()
	Finish()
}
```

下面是使用的例子:

> 场景一和场景二的例子,目前测试运行中,内存和cpu均稳定

### 调用例子一

代码见 examples/simple/example.go


```go
package main

import (
	"fmt"
	"gitlab.tools.vipshop.com/devops-team/go-pod/gopod"
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

```


### 调用例子二

这个例子,我们模拟一个秒级的调度的工作场景

代码见 examples/advance/example.go

```go
package main

import (
	"fmt"
	"gitlab.tools.vipshop.com/devops-team/go-pod/gopod"
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

	advancePod := gopod.NewPod(100) //协程池大小为100
	advancePod.Start()              //开启Pod调度

	go func() {
		for {
			select {
			case <-ticker:
				//每秒模拟派分500个作业
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
				os.Exit(2)
			}
		}
	}()

	time.Sleep(1 * time.Hour)
	exchan <- 1
}

```
