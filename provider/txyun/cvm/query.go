package cvm

import (
	"context"

	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"

	"github.com/zhou-lincong/CMDB/apps/host"
	"github.com/zhou-lincong/CMDB/utils"
)

//9.3	把查询到的腾讯云主机列表，存到CMDB里面的主机列表去
func (o *CVMOperator) Query(ctx context.Context, req *cvm.DescribeInstancesRequest) (
	*host.HostSet, error) {
	resp, err := o.client.DescribeInstancesWithContext(ctx, req)
	if err != nil {
		return nil, err
	}
	o.log.Debug(resp.ToJsonString()) //转成字符串
	// 	{"Response":{"TotalCount":1,"InstanceSet":[{"Placement":{"Zone":"ap-shanghai-2","ProjectId":0},"InstanceId":"ins-5aofqu5r","InstanceType":"SA2.MEDIUM2","CPU":2,"Memory":2,"RestrictState":"NORMAL","InstanceName":"未命名","InstanceChargeType":"SPOTPAID","SystemDisk":{"DiskType":"CLOUD_PREMIUM","DiskId":"disk-qter4hs1","DiskSize":50},"PrivateIpAddresses":["172.17.0.12"],"PublicIpAddresses":["1.116.131.189"],"InternetAccessible":{"InternetChargeType":"BANDWIDTH_POSTPAID_BY_HOUR","InternetMaxBandwidthOut":1},"VirtualPrivateCloud":{"VpcId":"vpc-1z0r0z40","SubnetId":"subnet-qt5e6nqt","AsVpcGateway":false},"ImageId":"img-eb30mz89","CreatedTime":"2022-05-26T10:17:34Z","OsName":"TencentOS Server 3.1 (TK4)","SecurityGroupIds":["sg-hvnd5t28"],"LoginSettings":{"KeyIds":["skey-n7gfqxmb"]},"InstanceState":"RUNNING","StopChargingMode":"NOT_APPLICABLE","Uuid":"2de0bd2d-a96f-4c2f-8a7c-c2314532d46d","DisasterRecoverGroupId":"","CamRoleName":"","HpcClusterId":"","Isolat
	// edSource":"NOTISOLATED"}],"RequestId":"852ab03c-850e-4afe-98fe-80b831db51f6"}}

	//数据查出来之后，转换成host.HostSet数据返回出去，才能入库到数据库到hsot--save
	//单独封装转换逻辑

	set := o.transferSet(resp.Response.InstanceSet)
	set.Total = utils.PtrInt64(resp.Response.TotalCount)

	return set, nil
}
