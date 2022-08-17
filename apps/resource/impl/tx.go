package impl

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zhou-lincong/CMDB/apps/resource"
)

//6.10 单独封装SaveResource逻辑
func SaveResource(ctx context.Context, tx *sql.Tx, base *resource.Base,
	info *resource.Information) error {
	////保存到resource表

	//避免sql注入，使用prepare
	stmt, err := tx.PrepareContext(ctx, sqlInsertResource)
	if err != nil {
		return fmt.Errorf("prepare insert resource error, %s", err)
	}
	defer stmt.Close()

	//保存资源数据，将数组IP[10.10.1.1,10.20.2.2]-->字符串10.10.1.1,10.20.2.2,去app.go单独封装转换逻辑
	_, err = stmt.ExecContext(ctx,
		base.Id, base.ResourceType, base.Vendor, base.Region, base.Zone, base.CreateAt, info.ExpireAt,
		info.Category, info.Type, info.Name, info.Description, info.Status, info.UpdateAt,
		base.SyncAt, info.SyncAccount, info.PublicIPToString(), info.PrivateIPToString(),
		info.PayType, base.DescribeHash, base.ResourceHash, base.SecretId, base.Domain,
		base.Namespace, base.Env, base.UsageMode,
	)

	////保存到resource_tag表
	if err := updateResourceTag(ctx, tx, base.Id, info.Tags); err != nil {
		return err
	}
	return nil
}

//6.12 封装updateResourceTag逻辑
func updateResourceTag(ctx context.Context, tx *sql.Tx, resourceId string, tags []*resource.Tag) error {
	stmt, err := tx.PrepareContext(ctx, sqlInsertOrUpdateResourceTag)
	if err != nil {
		return fmt.Errorf("prepare update resource tag error, %s", err)
	}
	defer stmt.Close()

	for i := range tags {
		t := tags[i]
		if t.Weight == 0 {
			t.Weight = 1
		}
		_, err = stmt.ExecContext(ctx,
			t.Type, t.Key, t.Value, t.Describe, resourceId, t.Weight, time.Now().UnixMilli(),
			t.Describe, t.Weight,
		)
		if err != nil {
			return fmt.Errorf("save resource tag error, %s", err)
		}
	}
	return nil
}

//6.16 封装SaveResource逻辑
func UpdateResource(ctx context.Context, tx *sql.Tx, base *resource.Base,
	info *resource.Information) error {
	//避免sql注入，使用prepare
	stmt, err := tx.PrepareContext(ctx, sqlUpdateResource)
	if err != nil {
		return fmt.Errorf("prepare update reousrce sql error, %s", err)
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx,
		info.ExpireAt, info.Category, info.Type, info.Name, info.Description,
		info.Status, info.UpdateAt, base.SyncAt, info.SyncAccount,
		info.PublicIPToString(), info.PrivateIPToString(), info.PayType, base.DescribeHash, base.ResourceHash,
		base.SecretId, base.Namespace, base.Env, base.UsageMode,
		base.Id,
	)
	if err != nil {
		return fmt.Errorf("update resource base info error, %s", err)
	}

	if err := updateResourceTag(ctx, tx, base.Id, info.Tags); err != nil {
		return fmt.Errorf("update resource tag error, %s", err)
	}

	return nil
}

func DeleteResource(ctx context.Context, tx *sql.Tx, id string) error {
	// 删除Resource表里面的数据
	stmt, err := tx.PrepareContext(ctx, slqDeleteResource)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, id)
	if err != nil {
		return err
	}

	// 删除Resource_Tag表里面的数据
	stmt, err = tx.PrepareContext(ctx, slqDeleteResourceTag)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, id)
	if err != nil {
		return err
	}

	return nil
}
