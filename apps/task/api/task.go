package api

import (
	"github.com/zhou-lincong/CMDB/apps/task"

	"github.com/emicklei/go-restful/v3"
	"github.com/infraboard/mcube/http/request"
	"github.com/infraboard/mcube/http/response"
)

func (h *handler) CreatTask(r *restful.Request, w *restful.Response) {
	req := task.NewCreateTaskRequst()
	if err := request.GetDataFromRequest(r.Request, req); err != nil {
		response.Failed(w, err)
		return
	}

	//如果请求使用 HTTP 基本身份验证，则 BasicAuth 返回请求的授权标头中提供的用户名和密码。
	r.Request.BasicAuth()
	// 直接启动一个goroutine 来执行,
	//但这种方式，任务的报错当代码跑到这里时，相当于return那里返回一个ok
	//不能知道当前任务的一个状态了，不能及时查到任务到底是成功还是失败
	//像http.go里面用postman测试的"stage": "SUCCESS",
	//由于是一个异步接口，状态是需要记录下来，就需要把这个task对象保存到数据库里面去方便查询
	//接下来才会轮到任务的查询和修改
	//但下面的逻辑没是办法传task.id下去的，task.id是创建完了自动生成的
	// 想要通过Task做异常, 这里需要改造, 支持传递Task Id 参数
	// go func() {
	// 	set, err := h.task.CreateTask(r.Request.Context(), req)
	// }()

	//这里的ctx来自r *restful.Request
	set, err := h.task.CreateTask(r.Request.Context(), req)
	if err != nil {
		response.Failed(w, err)
		return
	}

	response.Success(w, set)
}

func (h *handler) QueryTask(r *restful.Request, w *restful.Response) {
	// query := task.NewQueryTaskRequestFromHTTP(r.Request)
	// set, err := h.task.QueryTask(r.Request.Context(), query)
	// if err != nil {
	// 	response.Failed(w, err)
	// 	return
	// }

	response.Success(w, nil)
}

func (h *handler) DescribeTask(r *restful.Request, w *restful.Response) {
	// req := task.NewDescribeTaskRequestWithId(r.PathParameter("id"))
	// ins, err := h.task.DescribeTask(r.Request.Context(), req)
	// if err != nil {
	// 	response.Failed(w, err)
	// 	return
	// }

	response.Success(w, nil)
}
