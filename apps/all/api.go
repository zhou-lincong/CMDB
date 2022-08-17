package all

import (
	// 注册所有HTTP服务模块, 暴露给框架HTTP服务器加载
	_ "github.com/zhou-lincong/CMDB/apps/book/api"
	_ "github.com/zhou-lincong/CMDB/apps/host/api"
	_ "github.com/zhou-lincong/CMDB/apps/resource/api"
	_ "github.com/zhou-lincong/CMDB/apps/secret/api"
	_ "github.com/zhou-lincong/CMDB/apps/task/api"
)
