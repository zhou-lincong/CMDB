package impl_test

import (
	"context"
	"os"
	"testing"

	// 注册所有对象
	_ "github.com/zhou-lincong/CMDB/apps/all"
	"github.com/zhou-lincong/CMDB/apps/resource"
	"github.com/zhou-lincong/CMDB/apps/task"
	"github.com/zhou-lincong/CMDB/conf"

	"github.com/infraboard/mcube/app"
	"github.com/infraboard/mcube/logger/zap"
)

var (
	ins task.ServiceServer
)

func TestCreateTask(t *testing.T) {
	req := task.NewCreateTaskRequst()
	req.Type = task.Type_RESOURCE_SYNC
	req.Region = os.Getenv("TX_CLOUD_REGION")
	req.ResourceType = resource.Type_HOST
	req.SecretId = "cb41jm7j8ck1r969o3hg"
	taskIns, err := ins.CreateTask(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(taskIns)
	// 	=== RUN   TestCreateTask
	// 2022-07-09T11:05:44.475+0800	DEBUG	[secret]	impl/secret.go:109	sql: SELECT * FROM secret  WHERE id = ?  ;
	//     e:\goproject\CMDB\apps\task\impl\task_test.go:32: id:"cb4f127j8ck1g9144g70" secret_description:"袁鑫" data:{secret_id:"cb41jm7j8ck1r969o3hg" region:"ap-shanghai" timeout:1800} status:{stage:RUNNING start_at:1657335944477}
	// --- PASS: TestCreateTask (0.05s)
	// PASS
	// ok  	github.com/zhou-lincong/CMDB/apps/task/impl	0.851s
}

func init() {
	// 通过环境变量加载测试配置
	if err := conf.LoadConfigFromEnv(); err != nil {
		panic(err)
	}

	// 全局日志对象初始化
	zap.DevelopmentSetup()

	// 初始化所有实例
	if err := app.InitAllApp(); err != nil {
		panic(err)
	}

	ins = app.GetGrpcApp(task.AppName).(task.ServiceServer)
}
