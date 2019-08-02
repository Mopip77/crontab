package master

import (
	"encoding/json"
	"fmt"
	"github.com/Mopip77/crontab/common"
	"net"
	"net/http"
	"strconv"
	"time"
)

type ApiServer struct {
	httpServer *http.Server
}

var (
	G_apiServer *ApiServer
)

// 保存任务的接口
// POST job={"name": "job1", "command": "echo hello", "cronExpr": "* * * * *"}
func handleJobSave(resp http.ResponseWriter, req *http.Request) {
	// 任务保存到ETCD中

	var (
		err     error
		postJob string
		job     common.Job
		oldJob  *common.Job
		bytes   []byte
	)

	// 1.解析POST表单
	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	// 2.去表单的job字段
	postJob = req.PostForm.Get("job")
	// 3.反序列化
	if err = json.Unmarshal([]byte(postJob), &job); err != nil {
		goto ERR
	}

	if oldJob, err = G_jobMgr.SaveJob(&job); err != nil {
		goto ERR
	}

	// 返回正常
	if bytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
		resp.Write(bytes)
	}
	return

ERR:
	// 返回异常
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

// 删除任务接口
// POST /job/delete name=job1
func handleJobDelete(resp http.ResponseWriter, req *http.Request) {
	var (
		err           error
		deleteJobName string
		oldJob        *common.Job
		bytes         []byte
	)

	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	deleteJobName = req.PostForm.Get("name")
	if oldJob, err = G_jobMgr.DeleteJob(deleteJobName); err != nil {
		goto ERR
	}

	if bytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
		resp.Write(bytes)
	}

	return

ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

func handleJobList(resp http.ResponseWriter, req *http.Request) {
	var (
		jobList []*common.Job
		err     error
		bytes   []byte
	)

	if jobList, err = G_jobMgr.ListJob(); err != nil {
		goto ERR
	}

	// 返回正常
	if bytes, err = common.BuildResponse(0, "success", jobList); err == nil {
		resp.Write(bytes)
	}

	return
ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

func handleJobKill(resp http.ResponseWriter, req *http.Request) {
	var (
		err     error
		jobName string
		bytes   []byte
	)

	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	jobName = req.PostForm.Get("name")
	if err = G_jobMgr.KillJob(jobName); err != nil {
		goto ERR
	}

	if bytes, err = common.BuildResponse(0, "success", nil); err == nil {
		resp.Write(bytes)
	}

	return
ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

func handleJobLog(resp http.ResponseWriter, req *http.Request) {
	var (
		err error

		name       string
		skipParam  string
		limitParam string

		skip  int
		limit int

		logArr []*common.JobLog
		bytes  []byte
	)

	if err = req.ParseForm(); err != nil {
		goto ERR
	}

	name = req.Form.Get("name")
	skipParam = req.Form.Get("skip")
	limitParam = req.Form.Get("limit")

	if skip, err = strconv.Atoi(skipParam); err != nil {
		skip = 0
	}

	if limit, err = strconv.Atoi(limitParam); err != nil {
		limit = 20
	}

	if logArr, err = G_LogMgr.ListLog(name, skip, limit); err != nil {
		goto ERR
	}

	if bytes, err = common.BuildResponse(0, "success", logArr); err == nil {
		resp.Write(bytes)
	}

	return
ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

func handleWorkList(resp http.ResponseWriter, req *http.Request) {
	var (
		workerArr []string
		err       error
		bytes     []byte
	)

	if workerArr, err = G_workerMgr.ListWorkers(); err != nil {
		goto ERR
	}

	if bytes, err = common.BuildResponse(0, "success", workerArr); err == nil {
		resp.Write(bytes)
	}

	return
ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		resp.Write(bytes)
	}
}

// 初始化服务
func InitApiServer() (err error) {
	var (
		mux           *http.ServeMux
		listener      net.Listener
		httpServer    *http.Server
		staticDir     http.Dir     // 静态文件根目录
		staticHandler http.Handler // 静态文件的HTTP回调
	)
	// route
	mux = http.NewServeMux()
	mux.HandleFunc("/job/save", handleJobSave)
	mux.HandleFunc("/job/delete", handleJobDelete)
	mux.HandleFunc("/job/list", handleJobList)
	mux.HandleFunc("/job/kill", handleJobKill)
	mux.HandleFunc("/job/log", handleJobLog)
	mux.HandleFunc("/worker/list", handleWorkList)

	// static resource
	staticDir = http.Dir(G_config.WebRoot)
	staticHandler = http.FileServer(staticDir)
	// 例如访问/index.html mux先匹配上面的api都不匹配， 然后匹配下面的Handle，用的是前缀匹配，所以能匹配上
	mux.Handle("/", http.StripPrefix("/", staticHandler)) // 首先通过/匹配上，然后在StripPrefix去掉前缀/， 再交给staticHandler 处理

	// 启动TCP监听
	if listener, err = net.Listen("tcp", ":"+strconv.Itoa(G_config.ApiPort)); err != nil {
		fmt.Println(err)
		return
	}

	// 创建一个http服务
	httpServer = &http.Server{
		ReadTimeout:  time.Duration(G_config.ApiReadTimeout) * time.Millisecond,
		WriteTimeout: time.Duration(G_config.ApiWriteTimeout) * time.Millisecond,
		Handler:      mux,
	}

	// 赋值单例
	G_apiServer = &ApiServer{
		httpServer: httpServer,
	}

	// 启动
	go httpServer.Serve(listener)

	return
}
