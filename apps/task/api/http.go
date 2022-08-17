package api

import (
	"github.com/zhou-lincong/CMDB/apps/task"

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
	task task.ServiceServer
	log  logger.Logger
}

func (h *handler) Config() error {
	h.log = zap.L().Named(task.AppName)
	h.task = app.GetGrpcApp(task.AppName).(task.ServiceServer)
	return nil
}

func (h *handler) Name() string {
	return task.AppName
}

func (h *handler) Version() string {
	return "v1"
}

func (h *handler) Registry(ws *restful.WebService) {
	tags := []string{h.Name()}

	ws.Route(ws.POST("/").To(h.CreatTask).
		Doc("create a task").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Metadata(label.Resource, "task").
		Metadata(label.Action, label.Create.Value()).
		Metadata(label.Auth, label.Enable).
		Metadata(label.Permission, label.Enable).
		Reads(task.CreateTaskRequst{}).
		Writes(response.NewData(task.Task{})))

	// postman测试
	// {
	// 	"secret_id": "cb41jm7j8ck1r969o3hg",
	// 	"resource_type": "host",
	// 	"region": "ap-shanghai"
	//  }

	// {
	// 	"code": 0,
	// 	"data": {
	// 		"id": "cb43117j8ck2pn1nd2sg",
	// 		"secret_description": "袁鑫",
	// 		"data": {
	// 			"type": "RESOURCE_SYNC",
	// 			"dry_run": false,
	// 			"secret_id": "cb41jm7j8ck1r969o3hg",
	// 			"resource_type": "HOST",
	// 			"region": "ap-shanghai",
	// 			"params": {},
	// 			"timeout": 1800
	// 		},
	// 		"status": {
	// 			"stage": "SUCCESS",
	// 			"message": "",
	// 			"start_at": 1657286788049,
	// 			"end_at": 1657286788049,
	// 			"total_succeed": 0,
	// 			"total_failed": 0
	// 		}
	// 	}
	// }

	ws.Route(ws.GET("/").To(h.QueryTask).
		Doc("get all task").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Metadata(label.Resource, "task").
		Metadata(label.Action, label.List.Value()).
		Metadata(label.Auth, label.Enable).
		Metadata(label.Permission, label.Enable).
		Writes(response.NewData(task.TaskSet{})).
		Returns(200, "OK", task.TaskSet{}))

	ws.Route(ws.GET("/{id}").To(h.DescribeTask).
		Doc("describe an task").
		Param(ws.PathParameter("id", "identifier of the task").DataType("integer").DefaultValue("1")).
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Metadata(label.Resource, "task").
		Metadata(label.Action, label.Get.Value()).
		Metadata(label.Auth, label.Enable).
		Metadata(label.Permission, label.Enable).
		Writes(response.NewData(task.Task{})).
		Returns(200, "OK", response.NewData(task.Task{})).
		Returns(404, "Not Found", nil))
}

func init() {
	app.RegistryRESTfulApp(h)
}
