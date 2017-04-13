package gopod

//任务结构
type Task struct {
	TargetObj  *interface{}
	TargetFunc string
	CloseFunc  string
	StartTime  int64
	EndTime    int64
}

func CreateTask(targetObj interface{}, targetFunc string, closeFunc string) Task {
	t := Task{
		TargetObj:  &targetObj,
		TargetFunc: targetFunc,
		CloseFunc:  closeFunc,
	}
	return t
}
