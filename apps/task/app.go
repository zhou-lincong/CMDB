package task

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/rs/xid"
)

const (
	AppName = "task"
)

var (
	validate = validator.New()
)

func NewDefaultTask() *Task {
	return &Task{
		Data:   &CreateTaskRequst{},
		Status: &Status{},
	}
}

func (s *Task) Run() {
	s.Status.StartAt = time.Now().UnixMilli()
	s.Status.Stage = Stage_RUNNING
}

func NewCreateTaskRequst() *CreateTaskRequst {
	return &CreateTaskRequst{
		Params: map[string]string{},
		// 设置task的默认半小时的超时时间
		Timeout: 30 * 60,
	}
}

func (req *CreateTaskRequst) Validate() error {
	return validate.Struct(req)
}

func CreateTask(req *CreateTaskRequst) (*Task, error) {
	if err := req.Validate(); err != nil {
		return nil, err
	}

	ins := NewDefaultTask()
	ins.Data = req
	ins.Id = xid.New().String() //生成随机id

	return ins, nil
}

func (s *Task) Failed(message string) {
	s.Status.EndAt = time.Now().UnixMilli()
	s.Status.Stage = Stage_FAILED
	s.Status.Message = message
}

func (s *Task) Success() {
	s.Status.EndAt = time.Now().UnixMilli()
	s.Status.Stage = Stage_SUCCESS
}
