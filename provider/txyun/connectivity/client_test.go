package connectivity_test

import (
	"testing"

	"github.com/zhou-lincong/CMDB/provider/txyun/connectivity"
)

func TestTencentCloudClient(t *testing.T) {
	conn := connectivity.C()
	if err := conn.Check(); err != nil {
		t.Fatal(err)
	}
	t.Log(conn.AccountId())
}

func init() {
	//环境变量加载，赋值全局client,初始化client
	err := connectivity.LoadClientFromEnv()
	if err != nil {
		panic(err)
	}
}
