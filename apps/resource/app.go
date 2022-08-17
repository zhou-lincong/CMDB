//2.3	4.2
package resource

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/infraboard/mcube/http/request"
	"github.com/zhou-lincong/CMDB/utils"
)

const (
	//2.3
	AppName = "resource"
)

//4.2.1	判断请求是否携带Tag
func (r *SearchRequest) HasTag() bool {
	if len(r.Tags) > 0 {
		return true
	}
	return false
}

//4.2.2	定义Tag的比较操作符号，类比promethues
type Operator string

//4.2.2	定义Tag的比较操作符号，类比promethues
const (
	Operator_EQUAL          = "="  //SQL里面的比较操作：=
	Operator_NOT_EQUAL      = "!=" //SQL里面的比较操作：!=
	Operator_LIKE_EQUAL     = "=~" //SQL里面的比较操作：LIKE
	Operator_NOT_LIKE_EQUAL = "!~" //SQL里面的比较操作：NOT LIKE
)

//4.2.3	多个值比较的关系说明：
//	如果Tag是	app =~ app1,app2	这里app1和app2是OR的关系	是一种白名单策略(包含)
//	如果Tag是	app !~ app3,app4	这里app3和app4是AND的关系	是一种黑名单策略(排除)
func (s *TagSelector) RelationShip() string {
	switch s.Operater {
	case Operator_EQUAL, Operator_LIKE_EQUAL:
		return " OR "
	case Operator_NOT_EQUAL, Operator_NOT_LIKE_EQUAL:
		return " AND "
	default:
		return " OR "
	}
}

//4.2.4	ResourceSet列表构造函数
func NewResourceSet() *ResourceSet {
	return &ResourceSet{
		Items: []*Resource{},
	}
}

//4.2.4 	Resource构造函数
func NewDefaultResource() *Resource {
	return &Resource{
		Base:        &Base{},
		Information: &Information{},
	}
}

//4.2.5	将string格式处理成以“ ，”相隔的列表
func (i *Information) LoadPrivateIPString(s string) {
	if s != "" {
		i.PrivateIp = strings.Split(s, ",")
	}
}

//4.2.5	将string格式处理成以“ ，”相隔的列表
func (i *Information) LoadPublicIPString(s string) {
	if s != "" {
		i.PublicIp = strings.Split(s, ",")
	}
}

//4.2.6 为ResourceSet补充Add方法
func (s *ResourceSet) Add(item *Resource) {
	s.Items = append(s.Items, item)
}

//4.3.2 NewDefaultTag构造函数
func NewDefaultTag() *Tag {
	return &Tag{
		Type:   TagType_USER, //默认可以修改
		Weight: 1,
	}
}

//4.3.3 将当前resourceSet里所有的resourceId拿出来，在专门封装一个函数resourceId（）
func (s *ResourceSet) ResourceIds() (ids []string) {
	for i := range s.Items {
		ids = append(ids, s.Items[i].Base.Id)
	}
	return ids
}

//4.3.4 更新的逻辑：如果tag.resource_id==resource.id,就把tag加到resource Tags属性里面去
func (s *ResourceSet) UpdateTag(tags []*Tag) {
	for i := range tags {
		for j := range s.Items {
			if s.Items[j].Base.Id == tags[i].ResourceId {
				s.Items[j].Information.AddTag(tags[i])
			}
		}
	}
}

//4.3.5 添加AddTag方法
func (r *Information) AddTag(t *Tag) {
	r.Tags = append(r.Tags, t)
}

//5.3.1 构建NewSearchRequestFromHTTP
//keywords=xx&domain=xx&tag=app=~app1,app2,app3
func NewSearchRequestFromHTTP(r *http.Request) (*SearchRequest, error) {
	//先拿到query string
	qs := r.URL.Query() //
	req := &SearchRequest{
		Page:        request.NewPageRequestFromHTTP(r), //?
		Keywords:    qs.Get("keywords"),
		ExactMatch:  qs.Get("exact_match") == "true",
		Domain:      qs.Get("domain"),
		Namespace:   qs.Get("namespace"),
		Env:         qs.Get("env"),
		Status:      qs.Get("status"),
		SyncAccount: qs.Get("sync_account"),
		WithTags:    qs.Get("with_tags") == "true",
		Tags:        []*TagSelector{},
	}

	umStr := qs.Get("usage_mode")
	if umStr != "" {
		mode, err := ParseUsageModeFromString(umStr)
		if err != nil {
			return nil, err
		}
		req.UsageMode = &mode
	}

	rtStr := qs.Get("resource_type")
	if rtStr != "" {
		rt, err := ParseTypeFromString(rtStr)
		if err != nil {
			return nil, err
		}
		req.Type = &rt
	}

	//单独处理Tag参数 app ~= app1,app2,app3 -->TagSelector -->req
	tgStr := qs.Get("tag")
	if tgStr != "" {
		tg, err := NewTagsFromString(tgStr)
		if err != nil {
			return nil, err
		}
		req.AddTag(tg...)
	}

	return req, nil
}

//5.3.2 构建NewTagsFromString
//例key1=v1,v2,v3 & key2=~v1,v2,v3 ,前面是第一个selector，后面是and连接多个
func NewTagsFromString(tagStr string) (tags []*TagSelector, err error) {
	if tagStr == "" {
		return
	}
	//拿到tag之后，根据“&”符号进行拆分
	items := strings.Split(tagStr, "&")
	for _, v := range items {
		//key1=v1,v2,v3解析成-->TagSelector
		t, err := ParExpr(v)
		if err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}

//5.3.3 实现AddTag函数
func (req *SearchRequest) AddTag(t ...*TagSelector) {
	req.Tags = append(req.Tags, t...)
}

//5.3.4 实现ParExpr函数，解析
func ParExpr(str string) (*TagSelector, error) {
	op := ""
	kv := []string{}

	//app=~v1,v2,v3
	if strings.Contains(str, Operator_LIKE_EQUAL) {
		op = "LIKE" //
		kv = strings.Split(str, Operator_LIKE_EQUAL)
	} else if strings.Contains(str, Operator_NOT_LIKE_EQUAL) {
		op = "NOT LIKE"
		kv = strings.Split(str, Operator_NOT_LIKE_EQUAL)
	} else if strings.Contains(str, Operator_NOT_EQUAL) {
		op = "!="
		kv = strings.Split(str, Operator_NOT_EQUAL)
	} else if strings.Contains(str, Operator_EQUAL) {
		op = "="
		kv = strings.Split(str, Operator_EQUAL)
	} else {
		return nil, fmt.Errorf("no support operator [=, =~, !=, !~]")
	}

	if len(kv) != 2 {
		return nil, fmt.Errorf("key,value format error,requred key=value")
	}

	selector := &TagSelector{
		Key:      kv[0],
		Operater: op,
		Values:   []string{},
	}

	//v1,v2,v3 splite切开变成  [v1,v2,v3]
	//如果Value等于*表示只匹配key
	if kv[1] != "*" {
		selector.Values = strings.Split(kv[1], ",")
	}

	return selector, nil
}

//6.9 补充Information hash方法
func (i *Information) Hash() string {
	return utils.Hash(i)
}

//6.11 IPToString转换逻辑
func (i *Information) PrivateIPToString() string {
	return strings.Join(i.PrivateIp, ",")
}

//6.11 IPToString转换逻辑
func (i *Information) PublicIPToString() string {
	return strings.Join(i.PublicIp, ",")
}

//9.6
func NewThirdTag(key, value string) *Tag {
	return &Tag{
		Type:   TagType_THIRD,
		Key:    key,
		Value:  value,
		Weight: 1,
	}
}
