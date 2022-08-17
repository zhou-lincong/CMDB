//2.2	4.1
package impl

import (
	"context"
	"fmt"
	"strings"

	"github.com/infraboard/mcube/exception"
	"github.com/infraboard/mcube/sqlbuilder"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zhou-lincong/CMDB/apps/resource"
)

//2.2.1		4.1.1
func (s *service) Search(ctx context.Context, req *resource.SearchRequest) (
	*resource.ResourceSet, error) {
	//SQL是一个模板，到底应该使用左连接还是右连接，取决于是否需要关联Tag表
	//LEFT JOIN 是先扫描左表， RIGHT JOIN先扫描右表，当有Tag过滤，需要关联右表，可以以右表为准
	//如果扫描Tag表的成本比扫描Resource表的成本低，就可以使用RIGHT JOIN

	//默认是左连接
	join := "LEFT"
	//再判断请求里面右没有传Tag
	if req.HasTag() {
		join = "RIGHT" //如果有Tag，就设成右连接
	}
	//构建过滤条件
	builder := sqlbuilder.NewBuilder(fmt.Sprintf(sqlQueryResource, join))
	s.buildQuery(builder, req)

	////计数统计：COUNT语句
	set := resource.NewResourceSet()

	//获取total SELECT COUNT(*) FROM t Where ....
	countSQL, args := builder.BuildFromNewBase(fmt.Sprintf(sqlCountResource, join))
	countStmt, err := s.db.Prepare(countSQL)
	if err != nil {
		s.log.Debugf("count sql, %s, %v", countSQL, args)
		return nil, exception.NewInternalServerError("prepare count sql error, %s", err)
	}
	defer countStmt.Close()
	err = countStmt.QueryRow(args...).Scan(&set.Total)
	if err != nil {
		return nil, exception.NewInternalServerError("scan count value error,%s", err)
	}

	////查询分页数据：SELECT语句

	//tag查询时，以tag时间排序,如果没有Tag就以资源的创建时间进行排序
	//为什么要排序？因为希望最新添加的资源排在最前面
	//比如添加了一个资源，最后添加的资源，被最先看到，就是一个堆
	if req.HasTag() {
		builder.Order("t.create_at").Desc()
	} else {
		builder.Order("r.create_at").Desc()
	}

	//获取分页数据
	querySQL, args := builder.
		GroupBy("r.id").
		Limit(req.Page.ComputeOffset(), uint(req.Page.PageSize)).
		BuildQuery()
	s.log.Debugf("sql: %s, args: %v", querySQL, args)

	queryStmt, err := s.db.PrepareContext(ctx, querySQL)
	if err != nil {
		return nil, exception.NewInternalServerError("prepare query resource error, %s", err.Error())
	}
	defer queryStmt.Close()

	rows, err := queryStmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, exception.NewInternalServerError(err.Error())
	}
	defer rows.Close()

	var (
		publicIPList, privateIPList string
	)
	for rows.Next() {
		ins := resource.NewDefaultResource()
		base := ins.Base
		info := ins.Information
		err := rows.Scan(
			&base.Id, &base.ResourceType, &base.Vendor, &base.Region, &base.Zone, &base.CreateAt, &info.ExpireAt,
			&info.Category, &info.Type, &info.Name, &info.Description,
			&info.Status, &info.UpdateAt, &base.SyncAt, &info.SyncAccount,
			&publicIPList, &privateIPList, &info.PayType, &base.DescribeHash, &base.ResourceHash,
			&base.SecretId, &base.Domain, &base.Namespace, &base.Env, &base.UsageMode,
		)
		if err != nil {
			return nil, exception.NewInternalServerError("query resource error, %s", err.Error())
		}

		//存入数据库的是一个列表，格式：10.10.1.1，10.10.2.2，....
		//因此从数据库取出该数据，对格式进行特殊处理
		info.LoadPrivateIPString(privateIPList)
		info.LoadPublicIPString(publicIPList)
		set.Add(ins)
	}

	//补充资源的标签
	//为什么不在上个SQL，直接把Tsg查出来？
	//因为在SQL中查询Tag，只能查出匹配到的tag,如果用户传一个app=app1 就只有app=app1的标签，其他的标签就获取不到
	//因为where语句只匹配了这个标签，其他的标签就不会去取了，
	//如果想把所有的TAG查出来，就单独取通过资源维度去筛选这个标签
	if req.WithTags {
		tags, err := QueryTag(ctx, s.db, set.ResourceIds())
		if err != nil {
			return nil, err
		}

		//已经查询出这个set关联的所有Tag(resource_id)
		//对应resource的Tag更新到Resource结构体
		//更新的逻辑：如果tag.resource_id==resource.id,就把tag加到resource Tags属性里面去
		set.UpdateTag(tags)
	}

	//最后将set数据结构返回出去
	return set, nil
}

//4.1.2	单独封装Query sql builder
func (s *service) buildQuery(builder *sqlbuilder.Builder, req *resource.SearchRequest) {
	//参数里面右模糊匹配与关键字匹配
	if req.Keywords != "" {
		if req.ExactMatch {
			//精准匹配
			builder.Where("r.name =? OR r.id =? OR r.private_ip =? OR r.public_ip =?",
				req.Keywords, req.Keywords, req.Keywords, req.Keywords,
			)
		} else {
			//模糊匹配
			builder.Where("r.name LIKE ? OR r.id LIKE ? OR r.private_ip LIKE ? OR r.public_ip LIKE ?",
				"%"+req.Keywords+"%", "%"+req.Keywords+"%", req.Keywords+"%", req.Keywords+"%",
			)
		}
	}

	//按照资源属性过滤
	if req.Domain != "" {
		builder.Where("r.domain =?", req.Domain)
	}
	if req.Namespace != "" {
		builder.Where("r.namespace = ?", req.Namespace)
	}
	if req.Env != "" {
		builder.Where("r.env = ?", req.Env)
	}
	if req.UsageMode != nil {
		builder.Where("r.usage_mode = ?", req.UsageMode)
	}
	if req.Vendor != nil {
		builder.Where("r.vendor = ?", req.Vendor)
	}
	if req.SyncAccount != "" {
		builder.Where("r.sync_accout = ?", req.SyncAccount)
	}
	if req.Type != nil {
		builder.Where("r.resource_type = ?", req.Type)
	}
	if req.Status != "" {
		builder.Where("r.status = ?", req.Status)
	}

	//如何通过Tag匹配资源，通过tag key 和 tag value进行联表查询 再配上where条件
	//允许输入多个Tag来对资源进行解锁。多个Tag之间的关系，是AND还是OR 例app=v1,product=p2
	//实现的策略：基于AND
	for i := range req.Tags {
		selector := req.Tags[i]

		//如果tag：=v1,没有key，就不成立不让过滤。作为tag查询，key是必须的
		if selector.Key == "" {
			continue
		}

		//添加key的过滤条件,默认支持 .*=v1查询,将 .* 转成% 定制化key如何通配，可以支持正则，但是性能不好，还需要左全局索引，最好用LIKE
		builder.Where("t.t_key LIKE ?", strings.ReplaceAll(selector.Key, ".*", "%"))

		//添加value的过滤条件
		//场景1：定制value通配，例app=app1或app2或app3
		//tag_value=? OR tag_value=?,有几个Tag Value就需要构造几个Where OR条件
		//场景2：如果tag是一个带有比较符号，例app_count >1
		//	(tag_value LIKE ? OR tag_value LIKE ?)
		var (
			condtions []string
			args      []interface{}
		)
		for _, v := range selector.Values {
			//构造这样的表达式： t.t_value [ = != =~ !~ ] value
			condtions = append(condtions, fmt.Sprintf("t.t_value %s ?", selector.Operater))
			//条件参数	args = append(args,v)也可以
			//将 .*转成%	做的特殊处理，为了匹配正则里面的 .*
			args = append(args, strings.ReplaceAll(v, ".*", "%"))
		}

		//如果tag的value是由多个条件组成 例app=~app1,app2 根据表达式[= != =~ !~],来智能决定value之间的关系
		if len(condtions) > 0 {
			vwhere := fmt.Sprintf("( %s )", strings.Join(condtions, selector.RelationShip()))
			builder.Where(vwhere, args...)
		}

	}
}

//2.2.2
func (s *service) QueryTag(ctx context.Context, req *resource.QueryTagRequest) (
	*resource.TagSet, error) {
	return nil, status.Errorf(codes.Unimplemented, "method QueryTag not implemented")
}

//2.2.3
func (s *service) UpdateTag(ctx context.Context, req *resource.UpdateTagRequest) (
	*resource.Resource, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateTag not implemented")
}
