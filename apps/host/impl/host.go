package impl

import (
	"context"
	"database/sql"

	"github.com/infraboard/mcube/exception"
	"github.com/infraboard/mcube/sqlbuilder"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/zhou-lincong/CMDB/apps/host"
	"github.com/zhou-lincong/CMDB/apps/resource/impl"
)

//6.4   6.14
func (s *service) SyncHost(ctx context.Context, ins *host.Host) (*host.Host, error) {
	exits, err := s.DescribeHost(ctx, host.NewDescribeHostRequestWithID(ins.Base.Id))
	if err != nil {
		//如果不是not found则直接返回
		if !exception.IsNotFoundError(err) {
			return nil, err
		}
	}

	//检查ins已经存在 则需要更新ins
	if exits != nil {
		s.log.Debugf("update host : %s", ins.Base.Id)
		exits.Put(host.NewUpdateHostDataByIns(ins))
		if err := s.update(ctx, exits); err != nil {
			return nil, err
		}
		return ins, nil
	}

	s.log.Debugf("insert host: %s", ins.Base.Id)
	//如果没有则说明是个新对象，直接保存
	if err := s.save(ctx, ins); err != nil {
		return nil, err
	}

	return ins, nil
	//return nil, status.Errorf(codes.Unimplemented, "method SyncHost not implemented")
}

//6.4	7.1
func (s *service) QueryHost(ctx context.Context, req *host.QueryHostRequest) (
	*host.HostSet, error) {
	query := sqlbuilder.NewBuilder(queryHostSQL)

	if req.Keywords != "" {
		query.Where("r.name LIKE ? OR r.id =? OR r.instance_id =? OR r.private_ip LIKE ? OR r.public_ip LIKE ?",
			"%"+req.Keywords+"",
			req.Keywords,
			req.Keywords,
			req.Keywords+"%",
			req.Keywords+"%",
		)
	}

	set := host.NewHostSet()

	//获取totalSELECT COUNT(*) FROMT t where
	countSQL, args := query.BuildFromNewBase(countHostSQL)
	countStmt, err := s.db.PrepareContext(ctx, countSQL)
	s.log.Debugf("queryHost count sql : %s", countSQL)
	if err != nil {
		return nil, exception.NewInternalServerError(err.Error())
	}
	defer countStmt.Close()

	err = countStmt.QueryRowContext(ctx, args...).Scan(&set.Total)
	if err != nil {
		return nil, exception.NewInternalServerError(err.Error())
	}

	//获取分页数据
	querySQL, args := query.
		GroupBy("r.id").
		Order("r.sync_at").
		Desc().
		Limit(req.Page.ComputeOffset(), uint(req.Page.PageSize)).
		BuildQuery()
	s.log.Debugf("query sql: %s", querySQL)
	queryStmt, err := s.db.PrepareContext(ctx, querySQL)
	if err != nil {
		return nil, exception.NewInternalServerError("prepare query host error, %s", err.Error())
	}
	defer queryStmt.Close()

	rows, err := queryStmt.QueryContext(ctx, args...)
	if err != nil {
		return nil, exception.NewInternalServerError(err.Error())
	}
	defer rows.Close()

	var (
		publicIPList, privateIPList, keyPairNameList, securityGroupsList string
	)
	for rows.Next() {
		ins := host.NewDefaultHost()
		base := ins.Base
		info := ins.Information
		desc := ins.Describe
		err = rows.Scan(
			&base.Id, &base.ResourceType, &base.Vendor, &base.Region, &base.Zone, &base.CreateAt, &info.ExpireAt,
			&info.Category, &info.Type, &info.Name, &info.Description,
			&info.Status, &info.UpdateAt, &base.SyncAt, &info.SyncAccount,
			&publicIPList, &privateIPList, &info.PayType, &base.DescribeHash, &base.ResourceHash,
			&base.SecretId, &base.Domain, &base.Namespace, &base.Env, &base.UsageMode, &base.Id,
			&desc.Cpu, &desc.Memory, &desc.GpuAmount, &desc.GpuSpec, &desc.OsType, &desc.OsName,
			&desc.SerialNumber, &desc.ImageId, &desc.InternetMaxBandwidthOut, &desc.InternetMaxBandwidthIn,
			&keyPairNameList, &securityGroupsList,
		)
		if err != nil {
			return nil, exception.NewInternalServerError("query host error, %s", err.Error())
		}

		info.LoadPrivateIPString(privateIPList)
		info.LoadPublicIPString(publicIPList)
		desc.LoadKeyPairNameString(keyPairNameList)
		desc.LoadSecurityGroupsString(securityGroupsList)
		set.Add(ins)
	}

	tags, err := impl.QueryTag(ctx, s.db, set.ResourceIds())
	if err != nil {
		return nil, err
	}
	set.UpdateTag(tags)

	return set, nil
}

//6.4		6.17 以resource 表为准 连接了2张表：resource_host + resource_tag
func (s *service) DescribeHost(ctx context.Context, req *host.DescribeHostRequest) (
	*host.Host, error) {
	query := sqlbuilder.NewBuilder(queryHostSQL).GroupBy("r.id")
	cond, val := req.Where()
	querySQL, args := query.Where(cond, val).BuildQuery()
	s.log.Debugf("sql: %s", querySQL)

	queryStmt, err := s.db.PrepareContext(ctx, querySQL)
	if err != nil {
		return nil, exception.NewInternalServerError("prepare describe host error, %s", err.Error())
	}
	defer queryStmt.Close()

	ins := host.NewDefaultHost()

	var (
		publicIPList, privateIPList, keyPairNameList, securityGroupsList string
	)

	base := ins.Base
	info := ins.Information
	desc := ins.Describe
	err = queryStmt.QueryRowContext(ctx, args...).Scan(
		&base.Id, &base.ResourceType, &base.Vendor, &base.Region, &base.Zone, &base.CreateAt, &info.ExpireAt,
		&info.Category, &info.Type, &info.Name, &info.Description,
		&info.Status, &info.UpdateAt, &base.SyncAt, &info.SyncAccount,
		&publicIPList, &privateIPList, &info.PayType, &base.DescribeHash, &base.ResourceHash,
		&base.SecretId, &base.Domain, &base.Namespace, &base.Env, &base.UsageMode, &base.Id,
		&desc.Cpu, &desc.Memory, &desc.GpuAmount, &desc.GpuSpec, &desc.OsType, &desc.OsName,
		&desc.SerialNumber, &desc.ImageId, &desc.InternetMaxBandwidthOut, &desc.InternetMaxBandwidthIn,
		&keyPairNameList, &securityGroupsList,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, exception.NewNotFound("%#v not found", req)
		}
		return nil, exception.NewInternalServerError("describe host error, %s", err.Error())
	}

	info.LoadPrivateIPString(privateIPList)
	info.LoadPublicIPString(publicIPList)
	desc.LoadKeyPairNameString(keyPairNameList)
	desc.LoadSecurityGroupsString(securityGroupsList)

	return ins, nil
}

//6.4
func (s *service) UpdateHost(ctx context.Context, req *host.UpdateHostRequest) (
	*host.Host, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UpdateHost not implemented")
}

//6.4
func (s *service) ReleaseHost(ctx context.Context, req *host.ReleaseHostRequest) (
	*host.Host, error) {
	return nil, status.Errorf(codes.Unimplemented, "method ReleaseHost not implemented")
}
