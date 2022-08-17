package all

import (
	// 注册所有GRPC服务模块, 暴露给框架GRPC服务器加载, 注意 导入有先后顺序
	_ "github.com/zhou-lincong/CMDB/apps/book/impl"
	_ "github.com/zhou-lincong/CMDB/apps/host/impl"
	_ "github.com/zhou-lincong/CMDB/apps/resource/impl"
	_ "github.com/zhou-lincong/CMDB/apps/secret/impl"
	_ "github.com/zhou-lincong/CMDB/apps/task/impl"
)
