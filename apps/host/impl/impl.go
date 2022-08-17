//6.3
package impl

import (
	"database/sql"

	"github.com/zhou-lincong/CMDB/apps/host"
	"github.com/zhou-lincong/CMDB/conf"

	"github.com/infraboard/mcube/app"
	"github.com/infraboard/mcube/logger"
	"github.com/infraboard/mcube/logger/zap"
	"google.golang.org/grpc"
)

var (
	//Service服务实例
	svr = &service{}
)

type service struct {
	db  *sql.DB
	log logger.Logger
	//作为必须实现grpc service 必须嵌入这个结构体
	host.UnimplementedServiceServer
}

func (s *service) Name() string {
	return host.AppName
}

func (s *service) Config() error {
	db, err := conf.C().MySQL.GetDB()
	if err != nil {
		return err
	}

	s.log = zap.L().Named(s.Name())
	s.db = db
	return nil
}

func (s *service) Registry(server *grpc.Server) {
	host.RegisterServiceServer(server, svr)
}

func init() {
	app.RegistryGrpcApp(svr)
}
