package impl

import (
	"context"
	"fmt"

	"github.com/zhou-lincong/CMDB/apps/host"
	"github.com/zhou-lincong/CMDB/apps/secret"
	"github.com/zhou-lincong/CMDB/apps/task"
	"github.com/zhou-lincong/CMDB/provider/txyun/connectivity"
	"github.com/zhou-lincong/CMDB/provider/txyun/cvm"
)

func newSyncHostRequest(secret *secret.Secret, task *task.Task) *syncHostReqeust {
	return &syncHostReqeust{
		secret: secret,
		task:   task,
	}
}

type syncHostReqeust struct {
	secret *secret.Secret
	task   *task.Task
}

//没有返回err，以后记录到task对象上去
func (i *impl) syncHost(ctx context.Context, req *syncHostReqeust) {
	// 处理任务状态
	//任务结束后调一下
	// go routine里面记住 一定要捕获异常, 程序绷掉
	// go recover 只能捕获 当前Gorouine的panice
	defer func() {
		if err := recover(); err != nil {
			// panic 任务失败
			req.task.Failed(fmt.Sprintf("pannic, %v", err))
		} else {
			// 正常结束，没有报错，返回成功
			if !req.task.Status.Stage.IsIn(task.Stage_FAILED, task.Stage_WARNING) {
				req.task.Success() //修改状态为2
			}
			//这里不用i.insert，因为task在save之前就已经保存了
			if err := i.update(ctx, req.task); err != nil {
				i.log.Errorf("save task error, %s", err)
			}
		}
	}()

	fmt.Println(" 初始化腾讯cvm operator")
	// 只实现主机同步, 初始化腾讯cvm operator
	// NewTencentCloudClient client
	txConn := connectivity.NewTencentCloudClient(
		req.secret.Data.ApiKey,
		req.secret.Data.ApiSecret,
		req.task.Data.Region)

	cvmOp := cvm.NewCVMOperator(txConn.CvmClient())
	fmt.Println(" 初始化腾讯cvm operator2")
	// 因为要同步所有资源，需要分页查询
	pagger := cvm.NewPagger(float64(req.secret.Data.RequestRate), cvmOp)
	for pagger.Next() {
		set := host.NewHostSet()
		// 查询分页有错误 反应在Task上面
		if err := pagger.Scan(ctx, set); err != nil {
			fmt.Println("scan err: ", err)
			req.task.Failed(err.Error())
			return
		}
		// 保持该页数据到数据库, 同步时间时, 记录下日志
		for index := range set.Items {
			_, err := i.host.SyncHost(ctx, set.Items[index])
			if err != nil {
				i.log.Errorf("sync host error, %s", err)
				continue
			}
		}
	}
}
