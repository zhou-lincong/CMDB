package host

import (
	context "context"
	"net/http"
	"strings"
	"time"

	"github.com/infraboard/mcube/flowcontrol/tokenbucket"
	"github.com/infraboard/mcube/http/request"
	"github.com/zhou-lincong/CMDB/apps/resource"
	"github.com/zhou-lincong/CMDB/utils"
)

const (
	AppName = "host"
)

//6.7 封装host.GenHash方法
func (h *Host) GenHash() error {
	//hash resource
	h.Base.ResourceHash = h.Information.Hash()
	//hash describe
	h.Base.DescribeHash = utils.Hash(h.Describe)
	return nil
}

//6.13 describe.string格式处理,入库的时候用string，用逗号连接起来
func (d *Describe) KeyPairNameToString() string {
	return strings.Join(d.KeyPairName, ",")
}

//6.13 describe.string格式处理
func (d *Describe) SecurityGroupsToString() string {
	return strings.Join(d.SecurityGroups, ",")
}

//6.18.3 加载的时候就根据逗号拆开
func (d *Describe) LoadKeyPairNameString(s string) {
	if s != "" {
		d.KeyPairName = strings.Split(s, ",")
	}
}

//6.18.3 加载的时候就根据逗号拆开
func (d *Describe) LoadSecurityGroupsString(s string) {
	if s != "" {
		d.SecurityGroups = strings.Split(s, ",")
	}
}

//6.15.1  DescribeHostRequestWithID构造函数
func NewDescribeHostRequestWithID(id string) *DescribeHostRequest {
	return &DescribeHostRequest{
		DescribeBy: DescribeBy_HOST_ID,
		Value:      id,
	}
}

//6.15.2 NewUpdateHostDataByIns构造函数
func NewUpdateHostDataByIns(ins *Host) *UpdateHostData {
	return &UpdateHostData{
		Information: ins.Information,
		Describe:    ins.Describe,
	}
}

//6.15.3 更新host
func (h *Host) Put(req *UpdateHostData) {
	oldRH := h.Base.ResourceHash
	oldDH := h.Base.DescribeHash
	h.Information = req.Information
	h.Describe = req.Describe
	h.Information.UpdateAt = time.Now().UnixMilli()
	//更新完成后，重新生成一次hash
	h.GenHash()
	//然后进行比对，设置更新状态
	if h.Base.ResourceHash != oldRH {
		h.Base.ResourceHashChanged = true
	}
	if h.Base.DescribeHash != oldDH {
		h.Base.DescribeHashChanged = true
	}
}

//6.18.1
func (req *DescribeHostRequest) Where() (string, interface{}) {
	switch req.DescribeBy {
	default:
		return "r.id =?", req.Value
	}
}

//6.18.2
func NewDefaultHost() *Host {
	return &Host{
		Base: &resource.Base{
			ResourceType: resource.Type_HOST,
		},
		Information: &resource.Information{},
		Describe:    &Describe{},
	}
}

//7.1
func NewHostSet() *HostSet {
	return &HostSet{
		Items: []*Host{},
	}
}

//7.1 把具体类型改成泛型，然后再断言
func (s *HostSet) Add(items any) {
	s.Items = append(s.Items, items.(*Host))
}

//7.1	获取ResourceId
func (s *HostSet) ResourceIds() (ids []string) {
	for i := range s.Items {
		ids = append(ids, s.Items[i].Base.Id)
	}
	return
}

//7.1
func (s *HostSet) UpdateTag(tags []*resource.Tag) { //?
	for i := range tags {
		for j := range s.Items {
			if s.Items[j].Base.Id == tags[i].ResourceId {
				s.Items[j].Information.AddTag(tags[i])
			}
		}
	}
}

func NewQueryHostFromHTTP(r *http.Request) *QueryHostRequest {
	qs := r.URL.Query()
	page := request.NewPageRequestFromHTTP(r)
	kw := qs.Get("kewords")
	return &QueryHostRequest{
		Page:     page,
		Keywords: kw,
	}
}

//10.3
func (s *HostSet) Length() int64 {
	return int64(len(s.Items))
}

// 10.1 分页器定义
// for p.Next() {
// 	if err := p.Scan(set); err != nil {
// 		...
// 	}
// }
type Pagger interface {
	Next() bool
	SetPageSize(ps int64)
	Scan(context.Context, *HostSet) error
}

// RDS ...
// Scan(context.Context, *RDSSet) error

type Set interface {
	// 往Set里面添加元素, 任何类型都可以
	Add(any)
	// 当前的集合里面有多个元素
	Length() int64
}

// 抽象通用Pager
type PagerV2 interface {
	Next() bool
	//不能传any
	Scan(context.Context, Set) error
	Offset() int64
	SetPageSize(ps int64)
	SetRate(r float64)
	PageSize() int64
	PageNumber() int64
}

func NewBasePagerV2() *BasePagerV2 {
	return &BasePagerV2{
		hasNext:    true,
		tb:         tokenbucket.NewBucketWithRate(1, 1),
		pageNumber: 1,
		pageSize:   20,
	}
}

// 面向组合, 用他来实现一个模板, 除了Scan的其他方法都实现
// 把通用的参数抽象出来
type BasePagerV2 struct {
	// 令牌桶
	hasNext bool
	tb      *tokenbucket.Bucket

	// 控制分页的核心参数
	pageNumber int64
	pageSize   int64
}

func (p *BasePagerV2) Next() bool {
	// 等待分配令牌
	p.tb.Wait(1)

	return p.hasNext
}

func (p *BasePagerV2) Offset() int64 {
	return (p.pageNumber - 1) * p.pageSize
}

func (p *BasePagerV2) SetPageSize(ps int64) {
	p.pageSize = ps
}

func (p *BasePagerV2) PageSize() int64 {
	return p.pageSize
}

func (p *BasePagerV2) PageNumber() int64 {
	return p.pageNumber
}

func (p *BasePagerV2) SetRate(r float64) {
	p.tb.SetRate(r)
}

func (p *BasePagerV2) CheckHasNext(current int64) {
	// 可以根据当前一页是满页来决定是否有下一页
	if current < p.pageSize {
		p.hasNext = false
	} else {
		// 直接调整指针到下一页
		p.pageNumber++
	}
}
