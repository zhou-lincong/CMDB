package cvm

import (
	"context"

	"github.com/zhou-lincong/CMDB/apps/host"

	"github.com/infraboard/mcube/flowcontrol/tokenbucket"
	"github.com/infraboard/mcube/logger"
	"github.com/infraboard/mcube/logger/zap"
	cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
	tx_cvm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/cvm/v20170312"
)

// 10.1  速率rate: 5 req/s
func NewPagger(rate float64, op *CVMOperator) host.Pagger {
	p := &pagger{
		op:         op,
		hasNext:    true,
		pageNumber: 1,
		pageSize:   20,
		log:        zap.L().Named("CVM"),
		// 11 增加令牌桶限流
		tb: tokenbucket.NewBucketWithRate(rate, 3),
	}

	//
	p.req = tx_cvm.NewDescribeInstancesRequest()
	p.req.Limit = &p.pageSize
	p.req.Offset = p.offset()
	return p
}

//10.2
type pagger struct {
	req *cvm.DescribeInstancesRequest
	op  *CVMOperator
	log logger.Logger

	hasNext bool
	// 11 增加令牌桶
	tb *tokenbucket.Bucket

	// 控制分页的核心参数
	pageNumber int64
	pageSize   int64
}

//设置速率
func (p *pagger) SetRate(r float64) {
	p.tb.SetRate(r)
}

// 10.4 用户传入他自己想要的一页多少个
func (p *pagger) SetPageSize(ps int64) {
	p.pageSize = ps
}

//  10.4 根据分页参数来计算
func (p *pagger) offset() *int64 {
	offSet := (p.pageNumber - 1) * p.pageSize
	return &offSet
}

// 10.2 需要在请求数据是 计算出来(根据当前页是否满)
func (p *pagger) Next() bool {
	return p.hasNext //11.2为了测试令牌桶，这里可改成true
}

// 10.5 修改Req 执行真正的下一页的offset
func (p *pagger) nextReq() *cvm.DescribeInstancesRequest {
	//11.1 等待分配令牌
	p.tb.Wait(1)

	p.req.Offset = p.offset()
	p.req.Limit = &p.pageSize
	return p.req
}

// 10.2
func (p *pagger) Scan(ctx context.Context, set *host.HostSet) error {
	p.log.Debugf("query page: %d", p.pageNumber)
	hs, err := p.op.Query(ctx, p.nextReq())
	if err != nil {
		return err
	}

	// 把查询出来的数据赋值给set
	for i := range hs.Items {
		set.Add(set.Items[i])
	}

	// 可以根据当前一页是满页来决定是否有下一页
	if hs.Length() < p.pageSize {
		p.hasNext = false
	}

	// 直接调整指针到下一页
	p.pageNumber++
	return nil
}
