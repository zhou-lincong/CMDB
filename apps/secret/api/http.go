package api

import (
	"github.com/zhou-lincong/CMDB/apps/secret"

	restfulspec "github.com/emicklei/go-restful-openapi"
	"github.com/emicklei/go-restful/v3"
	"github.com/infraboard/mcube/app"
	"github.com/infraboard/mcube/http/label"
	"github.com/infraboard/mcube/http/response"
	"github.com/infraboard/mcube/logger"
	"github.com/infraboard/mcube/logger/zap"
)

var (
	h = &handler{}
)

type handler struct {
	service secret.ServiceServer
	log     logger.Logger
}

func (h *handler) Config() error {
	h.log = zap.L().Named(secret.AppName)
	h.service = app.GetGrpcApp(secret.AppName).(secret.ServiceServer)
	return nil
}

func (h *handler) Name() string {
	return secret.AppName
}

func (h *handler) Version() string {
	return "v1"
}

func (h *handler) Registry(ws *restful.WebService) {
	tags := []string{h.Name()}

	//路由修饰，类似/cmdb/api/v1/book,/cmdb/api/v1是框架生成，book是book.AppName生成
	//自动添加，避免url重名 Route /cmdb/api/v1/book/ 对应的方法是h.CreateBook
	// 需要装饰路由:   Route Path 不需要认证
	//这些meta信息还只是在路由上面，需要在路由上面把装饰的信息取出来，
	// 加上业务逻辑，然后才能完成业务的判断.通过在中间件把路由上面的信息获取出来
	ws.Route(ws.POST("/").To(h.CreateSecret).
		Doc("create a secret").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		// 添加一下标签做 路由装饰
		Metadata(label.Resource, h.Name()).
		Metadata(label.Action, label.Create.Value()).
		// 是否开启认证
		Metadata(label.Auth, label.Enable).
		// 是否开启鉴权
		Metadata(label.Permission, label.Enable).
		// 开启行为审计
		Metadata(label.Audit, label.Enable).
		// 基于用户属性的权限装饰, 未实现
		Metadata(label.Allow, "admin").
		Reads(secret.CreateSecretRequest{}).
		Writes(response.NewData(secret.Secret{})))

	ws.Route(ws.GET("/").To(h.QuerySecret).
		Doc("get all secret").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Metadata(label.Resource, h.Name()).
		Metadata(label.Action, label.List.Value()).
		Metadata(label.Auth, label.Enable).
		Metadata(label.Permission, label.Enable).
		Metadata(label.Audit, label.Enable).
		Reads(secret.QuerySecretRequest{}).
		Writes(response.NewData(secret.SecretSet{})).
		Returns(200, "OK", secret.SecretSet{}))

	ws.Route(ws.GET("/{id}").To(h.DescribeSecret).
		Doc("describe an secret").
		Param(ws.PathParameter("id", "identifier of the secret").DataType("integer").DefaultValue("1")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Metadata(label.Resource, h.Name()).
		Metadata(label.Action, label.Get.Value()).
		Metadata(label.Auth, label.Enable).
		Metadata(label.Permission, label.Enable).
		Metadata(label.Audit, label.Enable).
		Writes(response.NewData(secret.Secret{})).
		Returns(200, "OK", response.NewData(secret.Secret{})).
		Returns(404, "Not Found", nil))

	ws.Route(ws.DELETE("/{id}").To(h.DeleteSecret).
		Doc("delete a secret").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Metadata(label.Resource, h.Name()).
		Metadata(label.Action, label.Delete.Value()).
		Metadata(label.Auth, label.Enable).
		Metadata(label.Permission, label.Enable).
		Metadata(label.Audit, label.Enable).
		Param(ws.PathParameter("id", "identifier of the secret").DataType("string")))
}

func init() {
	app.RegistryRESTfulApp(h)
}
