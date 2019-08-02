package common

import (
	"context"
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"strings"
	"time"
)

type Job struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	CronExpr string `json:"cronExpr"`
}

// 任务调度计划
type JobSchedulePlan struct {
	Job      *Job
	Expr     *cronexpr.Expression
	NextTime time.Time
}

type JobExecuteInfo struct {
	Job        *Job
	PlanTime   time.Time
	RealTime   time.Time
	CancelCtx  context.Context
	CancelFunc context.CancelFunc
}

type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo
	Output      []byte
	Err         error
	StartTime   time.Time
	EndTime     time.Time
}

// http 接口API
type Response struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

// 变化事件
type JobEvent struct {
	EventType int //save, delete
	Job       *Job
}

// log
type JobLog struct {
	JobName      string `bson:"jobName" json:"jobName"`
	Command      string `bson:"command" json:"command"`
	Err          string `bson:"err" json:"err"`
	Output       string `bson:"output" json:"output"`
	PlanTime     int64  `bson:"planTime" json:"planTime"`
	ScheduleTime int64  `bson:"scheduleTime" json:"scheduleTime"`
	StartTime    int64  `bson:"startTime" json:"startTime"`
	EndTime      int64  `bson:"endTime" json:"endTime"`
}

// monga 查询 filter
type JobLogFilter struct {
	JobName string `bson:"jobName"`
}

type SortLogByStartTime struct {
	SortOrder int `bson:"startTime"` // -1
}

// log batch
type LogBatch struct {
	Logs []interface{} // 日志
}

func BuildResponse(errno int, msg string, data interface{}) (resp []byte, err error) {
	// 定义一个response
	var (
		response Response
	)

	response.Errno = errno
	response.Msg = msg
	response.Data = data

	resp, err = json.Marshal(response)
	return
}

func UnpackJob(value []byte) (ret *Job, err error) {
	var (
		job *Job
	)
	job = &Job{}

	if err = json.Unmarshal(value, job); err != nil {
		return
	}
	ret = job
	return
}

// 从etcd的job key中提取任务名
func ExtractJobName(jobKey string) string {
	return strings.TrimPrefix(jobKey, JOB_SAVE_DIR)
}

// 从etcd的killer key中提取任务名
func ExtractKillerName(jobKey string) string {
	return strings.TrimPrefix(jobKey, JOB_KILLER_DIR)
}

func ExtractWorkerIp(workerKey string) string {
	return strings.TrimPrefix(workerKey, JOB_WORKER_DIR)
}

// 任务变化事件 1——更新任务 2——删除任务
func BuildJobEvent(eventType int, job *Job) *JobEvent {
	return &JobEvent{
		EventType: eventType,
		Job:       job,
	}
}

// 构建执行计划
func BuildJobSchedulePlan(job *Job) (jobSchedulePlan *JobSchedulePlan, err error) {
	var (
		expr *cronexpr.Expression
	)
	// 解析
	if expr, err = cronexpr.Parse(job.CronExpr); err != nil {
		return
	}

	// 生成任务调度对象
	jobSchedulePlan = &JobSchedulePlan{
		Job:      job,
		Expr:     expr,
		NextTime: expr.Next(time.Now()),
	}

	return
}

func BuildJobExecuteInfo(jobSchedulePlan *JobSchedulePlan) (jobExecuteInfo *JobExecuteInfo) {
	jobExecuteInfo = &JobExecuteInfo{
		Job:      jobSchedulePlan.Job,
		PlanTime: jobSchedulePlan.NextTime,
		RealTime: time.Now(),
	}
	jobExecuteInfo.CancelCtx, jobExecuteInfo.CancelFunc = context.WithCancel(context.TODO())
	return
}
