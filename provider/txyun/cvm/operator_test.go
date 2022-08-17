package cvm_test

import (
	"context"
	"testing"

	"github.com/infraboard/mcube/logger/zap"
	tx_cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"

	"github.com/zhou-lincong/CMDB/apps/host"
	"github.com/zhou-lincong/CMDB/provider/txyun/connectivity"
	"github.com/zhou-lincong/CMDB/provider/txyun/cvm"
)

var op *cvm.CVMOperator

func TestQuery(t *testing.T) {
	req := tx_cvm.NewDescribeInstancesRequest()
	set, err := op.Query(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(set)
	// total:1  items:{base:{id:"ins-5aofqu5r"  vendor:TENCENT  region:"ap-shanghai"  zone:"ap-shanghai-2"  create_at:1653560254000}  information:{type:"SA2.MEDIUM2"  name:"未命名"  status:"RUNNING"  public_ip:"1.116.131.189"  private_ip:"172.17.0.12"  pay_type:"SPOTPAID"}  describe:{cpu:2  memory:2  os_name:"TencentOS Server 3.1 (TK4)"  serial_number:"2de0bd2d-a96f-4c2f-8a7c-c2314532d46d"  image_id:"img-eb30mz89"  internet_max_bandwidth_out:1  key_pair_name:"skey-n7gfqxmb"  security_groups:"sg-hvnd5t28"}}
}

func TestPaggerQuery(t *testing.T) {
	p := cvm.NewPagger(5, op)
	for p.Next() {
		set := host.NewHostSet()
		if err := p.Scan(context.Background(), set); err != nil {
			panic(err)
		}
		t.Log("page query result: ", set)
	}
	// === RUN   TestPaggerQuery
	// 2022-07-09T09:51:01.376+0800	DEBUG	[CVM]	cvm/pagger.go:81	query page: 1
	// --- FAIL: TestPaggerQuery (0.45s)
	// panic: [TencentCloudSDKError] Code=AuthFailure.SecretIdNotFound, Message=The SecretId is not found, please ensure that your SecretId is correct., RequestId=c20c3b7d-ac91-4aac-988c-0fca7c6e303f [recovered]
}

func TestPagerV2Query(t *testing.T) {
	p := cvm.NewPagerV2(op)
	for p.Next() {
		set := host.NewHostSet()
		if err := p.Scan(context.Background(), set); err != nil {
			panic(err)
		}
		t.Log("page query result: ", set)
		// === RUN   TestPagerV2Query
		// 2022-07-12T10:13:28.657+0800	DEBUG	[CVM]	cvm/pagerv2.go:42	query page: 1
		// --- FAIL: TestPagerV2Query (0.45s)
		// panic: [TencentCloudSDKError] Code=AuthFailure.SecretIdNotFound, Message=The SecretId is not found, please ensure that your SecretId is correct., RequestId=7500845e-1353-4e7f-b1fd-ebf31e896547 [recovered]
		// 	panic: [TencentCloudSDKError] Code=AuthFailure.SecretIdNotFound, Message=The SecretId is not found, please ensure that your SecretId is correct., RequestId=7500845e-1353-4e7f-b1fd-ebf31e896547
	}
}

func init() {
	//环境变量加载，赋值全局client,初始化client
	err := connectivity.LoadClientFromEnv()
	if err != nil {
		panic(err)
	}

	//初始化log,开发者环境
	zap.DevelopmentSetup()

	op = cvm.NewCVMOperator(connectivity.C().CvmClient())
}
