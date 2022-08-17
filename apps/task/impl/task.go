package impl

import (
	"context"
	"fmt"
	"time"

	"github.com/zhou-lincong/CMDB/apps/resource"
	"github.com/zhou-lincong/CMDB/apps/secret"
	"github.com/zhou-lincong/CMDB/apps/task"
	"github.com/zhou-lincong/CMDB/conf"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//13.2 创建任务的业务逻辑
func (i *impl) CreateTask(ctx context.Context, req *task.CreateTaskRequst) (*task.Task, error) {
	// 创建Task实例
	t, err := task.CreateTask(req)
	if err != nil {
		return nil, err
	}

	// 1. 查询secret
	s, err := i.secret.DescribeSecret(ctx, secret.NewDescribeSecretRequest(req.SecretId))
	if err != nil {
		return nil, err
	}
	//当查到secret之后，方便知道是用哪个key作为数据同步
	t.SecretDescription = s.Data.Description

	// 并解密api sercret
	if err := s.Data.DecryptAPISecret(conf.C().App.EncryptKey); err != nil {
		return nil, err
	}

	// 需要把Task 标记为Running, 修改一下Task对象的状态，扩展一个run方法
	t.Run() //状态1
	var taskCancel context.CancelFunc
	switch req.Type {
	case task.Type_RESOURCE_SYNC:
		// 根据secret所属的厂商, 初始化对应厂商的operator
		switch s.Data.Vendor {
		case resource.Vendor_TENCENT:
			// 操作哪种资源:
			switch req.ResourceType {
			case resource.Type_HOST:
				//方式1：
				// 只实现主机同步, 初始化腾讯cvm operator
				// NewTencentCloudClient client
				// txConn := connectivity.NewTencentCloudClient(
				// 	s.Data.ApiKey, s.Data.ApiSecret, req.Region)
				// cvmOp := cvm.NewCVMOperator(txConn.CvmClient())

				// 因为要同步所有资源，需要分页查询
				// pagger := cvm.NewPagger(float64(s.Data.RequestRate), cvmOp)
				// for pagger.Next() {
				// 	set := host.NewHostSet()
				// 	// 查询分页有错误立即返回
				// 	if err := pagger.Scan(ctx, set); err != nil {
				// 		return nil, err
				// 	}
				// 	// 保持该页数据到数据库, 同步时间时, 记录下日志
				// 	for index := range set.Items {
				// 		_, err := i.host.SyncHost(ctx, set.Items[index])
				// 		if err != nil {
				// 			i.log.Errorf("sync host error, %s", err)
				// 			continue
				// 		}
				// 	}
				// }

				//方式2：
				// 直接使用goroutine 把最耗时的逻辑封装到后台运行
				//不能直接在i.syncHost里面传ctx传递Http 的ctx
				//ctx是控制goroutine的上下文，当ctx宕了之后，请求就结束了，然后goroutine退出
				//这里的ctx来自于上层http请求传进来的，来自api/task.go的h.task.CreateTask里面
				//因为要在后台运行，不能请求结束导致goroutine退出
				//所以这里单独给一个超时控制的ctx
				taskExecCtx, cancel := context.WithTimeout(
					context.Background(),
					time.Duration(req.Timeout)*time.Second,
				)
				// //这个cancel给到task上去，但task解决不了cancel的问题，除非保存在内存当中
				// //在实例类上面保存所有在后台运行的task，然后给它一个cancel
				// //可以动态的根据id来cancel掉context
				taskCancel = cancel
				go i.syncHost(taskExecCtx, newSyncHostRequest(s, t))

			case resource.Type_RDS:
				//账单
			case resource.Type_BILL:
			default:
				fmt.Println("不支持这种resource type:", req.ResourceType)
			}
		case resource.Vendor_ALIYUN:

		case resource.Vendor_HUAWEI:

		case resource.Vendor_AMAZON:

		case resource.Vendor_VSPHERE:
		default:
			return nil, fmt.Errorf("unknow resource type: %s", s.Data.Vendor)
		}

		// 2. 利用secret的信息, 初始化一个operater
		// 使用operator进行资源的操作, 比如同步

		// 调用host service 把数据入库
	case task.Type_RESOURCE_RELEASE:
	default:
		return nil, fmt.Errorf("unknow task type: %s", req.Type)
	}

	// 需要保存到数据库,保存失败就取消任务
	//这里的保存逻辑和上面goroutine的保存逻辑是同步进行的，可能存在冲突
	//除非在这里传个消息，等待上面goroutine做完才去更新，这才是合理的
	//或者在goroutine就不更新，把里面的defer挪到这里来做
	//理论上这里的逻辑是要比goroutine里面的快，因为goroutine里面还要去查secret查云商
	//可以考虑放一个无缓冲管道进行控制
	if err := i.insert(ctx, t); err != nil {
		if taskCancel != nil {
			taskCancel()
		}
		return nil, err
	}
	return t, nil
}

func (i *impl) QueryBook(context.Context, *task.QueryTaskRequest) (*task.TaskSet, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryBook not implemented")
}

func (i *impl) DescribeBook(context.Context, *task.DescribeTaskRequest) (*task.Task, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DescribeBook not implemented")
}
