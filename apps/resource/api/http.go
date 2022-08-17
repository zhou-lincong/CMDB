//5.1
package api

import (
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	"github.com/infraboard/mcube/app"
	"github.com/infraboard/mcube/http/label"
	"github.com/infraboard/mcube/http/response"
	"github.com/infraboard/mcube/logger"
	"github.com/infraboard/mcube/logger/zap"

	"github.com/zhou-lincong/CMDB/apps/resource"
)

var (
	h = &handler{}
)

type handler struct {
	service resource.ServiceServer
	log     logger.Logger
}

func (h *handler) Config() error {
	h.log = zap.L().Named(resource.AppName)
	h.service = app.GetGrpcApp(resource.AppName).(resource.ServiceServer)
	return nil
}

func (h *handler) Name() string {
	return resource.AppName
}

//Restful API version
//	/cmdb/api/v1/resource path 包含 API Version
//通过Version函数定义
func (h *handler) Version() string {
	return "v1"
}

func (h *handler) Registry(ws *restful.WebService) {
	tags := []string{h.Name()}

	//Restful API,resource=cmdb_resource,action:list,auth:true
	ws.Route(ws.GET("/search").To(h.SearchResource).
		//用于生成Swagger文档
		Doc("get all resources").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		//label用于做权限系统，只是一些标签，中间件的逻辑需要额外补充
		Metadata(label.Resource, h.Name()).               //装饰这个资源属于哪个
		Metadata(label.Action, label.List.Value()).       //进行list操作
		Metadata(label.Auth, label.Enable).               //auth需要被认证
		Reads(resource.SearchRequest{}).                  //请求的结构体是什么
		Writes(response.NewData(resource.ResourceSet{})). //返回的对象是什么
		Returns(200, "OK", resource.ResourceSet{}))

	//资源标签管理
	// ws.Route(ws.POST("/").To(h.AddTag).
	// 	Doc("add resource tags").
	// 	Metadata(restfulspec.KeyOpenAPITags, tags).
	// 	Metadata(label.ResourceLableKey, "tags").
	// 	Metadata(label.ActionLableKey, label.Create.Value()).
	// 	Metadata(label.AuthLabelKey, label.Create.Value()).
	// 	Reads([]*resource.Tag{}).
	// 	Writes(response.NewData(resource.ResourceSet{})))

	// ws.Route(ws.DELETE("/").To(h.RemoveTag).
	// 	Doc("remove resource tags").
	// 	Metadata(restfulspec.KeyOpenAPITags, tags).
	// 	Metadata(label.ResourceLableKey, "tags").
	// 	Metadata(label.ActionLableKey, label.Delete.Value()).
	// 	Metadata(label.AuthLabelKey, label.Enable).
	// 	Reads([]*resource.Tag{}).
	// 	Writes(response.NewData(resource.Resource{})))

	//资源发现
	// ws.Route(ws.GET("/discovery/prometheus").To(h.DiscoveryPrometheus).
	// 	Doc("discovery resource for prometheus").
	// 	Metadata(restfulspec.KeyOpenAPITags, tags).
	// 	Metadata(label.ResourceLableKey("prometheus_resource")).
	// 	Metadata(label.ActionLableKey, label.List.Value()).
	// 	Reads(resource.SearchRequest{}).
	// 	Writes(response.NewData(resource.ResourceSet{})).
	// 	Returns(200, "OK", resource.ResourceSet{}))
}

func init() {
	app.RegistryRESTfulApp(h)
}
