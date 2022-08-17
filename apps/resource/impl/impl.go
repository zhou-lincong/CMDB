//2.1
package impl

import (
	"database/sql"

	"github.com/infraboard/mcube/app"
	"github.com/infraboard/mcube/logger"
	"github.com/infraboard/mcube/logger/zap"
	"google.golang.org/grpc"

	"github.com/zhou-lincong/CMDB/apps/resource" //grpc pb文件
	"github.com/zhou-lincong/CMDB/conf"
)

var (
	//Service服务实例
	svr = &service{}
)

type service struct {
	db  *sql.DB       //
	log logger.Logger //依赖一个logger记录日志

	resource.UnimplementedServiceServer //作为必须实现grpc service 必须嵌入这个结构体
}

//服务的配置
func (s *service) Config() error {
	//数据库实例从全局配置里面拿
	db, err := conf.C().MySQL.GetDB()
	if err != nil {
		return err
	}
	//初始化一个log，log名字跟服务名字一样
	s.log = zap.L().Named(s.Name())
	s.db = db
	return nil
}

//服务的名称
func (s *service) Name() string {
	return resource.AppName
}

//服务注册的方法
func (s *service) Registry(server *grpc.Server) {
	resource.RegisterServiceServer(server, svr)
}

func init() {
	app.RegistryGrpcApp(svr)
}
