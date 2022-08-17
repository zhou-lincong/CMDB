package connectivity

import (
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	"github.com/zhou-lincong/CMDB/utils"

	//凭证管理服务，管理认证、管理身份信息的服务，临时访问
	sts "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sts/v20180813"
)

type TencentCloudClient struct {
	Region    string `env:"TX_CLOUD_REGION"`
	SecretID  string `env:"TX_CLOUD_SECRET_ID"`
	SecretKey string `env:"TX_CLOUD_KEY"`

	accountId string
	cvmConn   *cvm.Client
}

func NewTencentCloudClient(secretID, secretKey, region string) *TencentCloudClient {
	return &TencentCloudClient{
		Region:    region,
		SecretID:  secretID,
		SecretKey: secretKey,
	}
}

//生成全局client
var client *TencentCloudClient

//通过环境变量加载
func LoadClientFromEnv() error {
	client = &TencentCloudClient{}
	if err := env.Parse(client); err != nil {
		return err
	}
	return nil
}

//通过此函数初始化并获得client
func C() *TencentCloudClient {
	if client == nil {
		panic("please load config first")
	}
	return client
}

//UseCvmClient cvm
func (me *TencentCloudClient) CvmClient() *cvm.Client {
	if me.cvmConn != nil {
		return me.cvmConn
	}

	credential := common.NewCredential(
		me.SecretID, me.SecretKey,
	)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.ReqTimeout = 300 //5分钟超时
	cpf.Language = "en-US"

	cvmConn, _ := cvm.NewClient(credential, me.Region, cpf)
	me.cvmConn = cvmConn
	return me.cvmConn
}

//	获取客户端账号ID
func (me *TencentCloudClient) Check() error {
	credential := common.NewCredential(
		me.SecretID, me.SecretKey,
	)

	cpf := profile.NewClientProfile()
	cpf.HttpProfile.ReqMethod = "POST"
	cpf.HttpProfile.ReqTimeout = 300
	cpf.Language = "en-US"

	stsConn, _ := sts.NewClient(credential, me.Region, cpf)

	req := sts.NewGetCallerIdentityRequest()
	resp, err := stsConn.GetCallerIdentity(req)
	if err != nil {
		return fmt.Errorf("unable to initialize the STS client: %#v", err)
	}

	//resp.Response.AccountId是一个指针对象，封装utils.PtrStrV处理指针转成string
	//不能直接写me.accountId =*resp.Response.AccountId，空指针会报错
	me.accountId = utils.PtrStrV(resp.Response.AccountId)
	return nil
}

//9.2
func (me *TencentCloudClient) AccountId() string {
	return me.accountId
}
