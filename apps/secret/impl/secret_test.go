package impl_test

import (
	"context"
	"os"
	"testing"

	// 注册所有
	_ "github.com/zhou-lincong/CMDB/apps/all"
	"github.com/zhou-lincong/CMDB/apps/secret"
	"github.com/zhou-lincong/CMDB/conf"

	"github.com/infraboard/mcube/app"
	"github.com/infraboard/mcube/logger/zap"
)

var (
	ins secret.ServiceServer
)

func TestQuerySecret(t *testing.T) {
	ss, err := ins.QuerySecret(context.Background(), secret.NewQuerySecretRequest())
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ss)
	// 	=== RUN   TestQuerySecret
	// 2022-07-08T19:53:27.353+0800	DEBUG	[secret]	impl/secret.go:58	sql: SELECT * FROM secret  ORDER BY create_at DESC LIMIT ?,? ;, args: [0 20]
	//     e:\goproject\CMDB\apps\secret\impl\secret_test.go:26: total:1 items:{id:"cb41jm7j8ck1r969o3hg" create_at:1657280984492 data:{description:"袁鑫" allow_regions:"*" api_key:"woaini" api_secret:"******" request_rate:5}}
	// --- PASS: TestQuerySecret (0.01s)
	// PASS
	// ok  	github.com/zhou-lincong/CMDB/apps/secret/impl	0.664s
}

func TestDescribeSecret(t *testing.T) {
	ss, err := ins.DescribeSecret(context.Background(),
		secret.NewDescribeSecretRequest("cb41jm7j8ck1r969o3hg"))
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ss)
	// 	=== RUN   TestDescribeSecret
	// 2022-07-08T19:53:03.290+0800	DEBUG	[secret]	impl/secret.go:109	sql: SELECT * FROM secret  WHERE id = ?  ;
	//     e:\goproject\CMDB\apps\secret\impl\secret_test.go:41: id:"cb41jm7j8ck1r969o3hg" create_at:1657280984492 data:{description:"袁鑫" allow_regions:"*" api_key:"woaini" api_secret:"@ciphered@p5KLoEK/Up6utPRLyks8hUHu0Cg6bvpuGgl5s0pisIE=" request_rate:5}
	// --- PASS: TestDescribeSecret (0.00s)
	// PASS
	// ok  	github.com/zhou-lincong/CMDB/apps/secret/impl	0.664s
}

func TestCreateSecret(t *testing.T) {
	req := secret.NewCreateSecretRequest()
	req.Description = "袁鑫"
	req.ApiKey = os.Getenv("TX_CLOUD_SECRET_ID")
	req.ApiSecret = os.Getenv("TX_CLOUD_KEY")
	req.AllowRegions = []string{"*"}
	ss, err := ins.CreateSecret(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(ss)
	// === RUN   TestCreateSecret
	//     e:\goproject\CMDB\apps\secret\impl\secret_test.go:53: id:"cb41c9vj8ck2u820qdkg" create_at:1657280039928 data:{description:"测试用例" allow_regions:"*" api_key:"AKIDb1B8oDLDo79tNaXDckbBUWeV4ORDhvrW" api_secret:"@ciphered@fW5SKf6pYRoR7npZw5kCnl7eX4RJ4+zyB4TS/NgPesidPe4jSRqH5TG9GSnFAUwLgh77wcT+2Ftzv3yLwl3bDA==" request_rate:5}
	// --- PASS: TestCreateSecret (0.08s)
	// PASS
	// ok  	github.com/zhou-lincong/CMDB/apps/secret/impl	0.818s
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

	ins = app.GetGrpcApp(secret.AppName).(secret.ServiceServer)

}
