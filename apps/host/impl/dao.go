package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zhou-lincong/CMDB/apps/host"
	"github.com/zhou-lincong/CMDB/apps/resource/impl"
)

//6.6
func (s *service) save(ctx context.Context, h *host.Host) error {
	//添加创建时间
	if h.Base.SyncAt != 0 {
		h.Base.SyncAt = time.Now().UnixMilli()
	}

	var (
		stmt *sql.Stmt
		err  error
	)

	//开启一个事务
	//文档请参考：http://cngolib.com/database-sql.html#db-begintx
	//关于事务级别可以参考文章：https://zhuanlan.zhihu.com/p/117476959
	//wiki: https://en.wikipedia.org/wiki/Isolation_(database_systems)#Isolation_levels
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("start save tx error, %s", err)
	}

	//执行结果提交或者回滚事务
	//当使用sql.Tx的操作方式操作数据后，需要使用sql.Tx的Commit方法显式的提交事务
	//如果出错，可以使用sql.Tx的Rollback方法回滚事务，保持数据的一致性
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				s.log.Errorf("save rollback error, %s", err)
			}
		} else {
			if err := tx.Commit(); err != nil {
				s.log.Errorf("save commimt error,%s", err)
			}
		}
	}()

	//生成描写信息的Hash，分别计算Resource Information Hash和Describe(host独有属性)的Hash
	//有专门把一个对象-->hash库，这里不选择这么做
	//把对象-->json(string)-->hash(string)-->对象
	if err := h.GenHash(); err != nil {
		return err
	}

	//保存资源基础信息(公共信息) 这里去resource里面单独封装SaveResource逻辑
	err = impl.SaveResource(ctx, tx, h.Base, h.Information)
	if err != nil {
		return err
	}

	//避免SQL注入，使用prepare
	stmt, err = tx.PrepareContext(ctx, insertHostSQL)
	if err != nil {
		return fmt.Errorf("prepare insert host sql error, %s", err)
	}
	defer stmt.Close()

	desc := h.Describe
	_, err = stmt.ExecContext(ctx,
		h.Base.Id, desc.Cpu, desc.Memory, desc.GpuAmount, desc.GpuSpec, desc.OsType, desc.OsName,
		desc.SerialNumber, desc.ImageId, desc.InternetMaxBandwidthOut,
		desc.InternetMaxBandwidthIn, desc.KeyPairNameToString(), desc.SecurityGroupsToString(),
	)
	if err != nil {
		return fmt.Errorf("save host resource describe error, %s", err)
	}

	return nil
}

func (s *service) update(ctx context.Context, ins *host.Host) error {
	var (
		stmt *sql.Stmt
		err  error
	)

	//开启一个事务
	//文档请参考：http://cngolib.com/database-sql.html#db-begintx
	//关于事务级别可以参考文章：https://zhuanlan.zhihu.com/p/117476959
	//wiki: https://en.wikipedia.org/wiki/Isolation_(database_systems)#Isolation_levels
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("start update tx error, %s", err)
	}

	//执行结果提交或者回滚事务
	//当使用sql.Tx的操作方式操作数据后，需要使用sql.Tx的Commit方法显式的提交事务
	//如果出错，可以使用sql.Tx的Rollback方法回滚事务，保持数据的一致性
	defer func() {
		if err != nil {
			if err := tx.Rollback(); err != nil {
				s.log.Errorf("update rollback error, %s", err)
			}
		} else {
			if err := tx.Commit(); err != nil {
				s.log.Errorf("update commimt error,%s", err)
			}
		}
	}()

	//更新资源基本信息
	if ins.Base.DescribeHashChanged {
		if err := impl.UpdateResource(ctx, tx, ins.Base, ins.Information); err != nil {
			return err
		}
	} else {
		s.log.Debug("resource data hash not changed, needn't update")
	}

	//更新实例信息
	if ins.Base.DescribeHashChanged {
		stmt, err = tx.PrepareContext(ctx, updateHostSQL)
		if err != nil {
			return fmt.Errorf("prepare update host sql error, %s", err)
		}
		defer stmt.Close()

		base := ins.Base
		desc := ins.Describe
		_, err = stmt.ExecContext(ctx,
			desc.Cpu, desc.Memory, desc.GpuAmount, desc.GpuSpec, desc.OsType, desc.OsName,
			desc.ImageId, desc.InternetMaxBandwidthOut,
			desc.InternetMaxBandwidthIn, desc.KeyPairNameToString(), desc.SecurityGroupsToString(),
			base.Id,
		)
		if err != nil {
			return fmt.Errorf("update host resource describe error, %s", err)
		}
	} else {
		s.log.Debug("describe data hash not changed, needn't update")
	}

	return nil
}
