package worker

import (
	"fmt"
	"github.com/Mopip77/crontab/common"
	"time"
)

type Scheduler struct {
	jobEventChan     chan *common.JobEvent
	jobPlanTable     map[string]*common.JobSchedulePlan
	jobExcutingTable map[string]*common.JobExecuteInfo
	jobResultChan    chan *common.JobExecuteResult
}

var (
	G_Scheduler *Scheduler
)

// 处理任务事件
func (scheduler *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	var (
		jobSchedulePlan *common.JobSchedulePlan
		jobExecuteInfo  *common.JobExecuteInfo
		jobExecuting    bool
		err             error
		jobExisted      bool
	)
	switch jobEvent.EventType {
	case common.JOB_EVENT_SAVE:
		if jobSchedulePlan, err = common.BuildJobSchedulePlan(jobEvent.Job); err != nil {
			return
		}
		scheduler.jobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
	case common.JOB_EVENT_DELETE:
		if jobSchedulePlan, jobExisted = scheduler.jobPlanTable[jobEvent.Job.Name]; jobExisted {
			delete(scheduler.jobPlanTable, jobEvent.Job.Name)
		}
	case common.JOB_EVENT_KILL:
		// 取消command执行
		if jobExecuteInfo, jobExecuting = scheduler.jobExcutingTable[jobEvent.Job.Name]; jobExecuting {
			jobExecuteInfo.CancelFunc() // 触发command杀死子进程
		}
	}
}

// 处理任务执行结果
func (scheduler *Scheduler) handleJobResult(result *common.JobExecuteResult) {
	var (
		jobLog *common.JobLog
	)
	// 删除excute表中的记录
	delete(scheduler.jobExcutingTable, result.ExecuteInfo.Job.Name)

	// log
	if result.Err != common.ERR_LOCK_ALRERADY_REQUIRED {
		jobLog = &common.JobLog{
			JobName:      result.ExecuteInfo.Job.Name,
			Command:      result.ExecuteInfo.Job.Command,
			Output:       string(result.Output),
			PlanTime:     result.ExecuteInfo.PlanTime.UnixNano() / 1e6,
			ScheduleTime: result.ExecuteInfo.RealTime.UnixNano() / 1e6,
			StartTime:    result.StartTime.UnixNano() / 1e6,
			EndTime:      result.EndTime.UnixNano() / 1e6,
		}
		if result.Err != nil {
			jobLog.Err = result.Err.Error()
		} else {
			jobLog.Err = ""
		}
		G_logSink.Append(jobLog)
	}

	//fmt.Println("任务执行完成：", result.ExecuteInfo.Job.Name, string(result.Output), result.Err)
}

// 尝试执行任务
func (scheduler *Scheduler) TryStartJob(jobSchedulePlan *common.JobSchedulePlan) {
	// 调度和执行是2件事情
	// 任务可能执行1分钟但是每秒调度一次， 即调度60次， 执行一次， 防止并发

	var (
		jobExecuteInfo *common.JobExecuteInfo
		jobExecuting   bool
	)

	// 如果任务正在执行， 跳过本次执行
	if jobExecuteInfo, jobExecuting = scheduler.jobExcutingTable[jobSchedulePlan.Job.Name]; jobExecuting {
		return
	}

	// 构建执行信息
	jobExecuteInfo = common.BuildJobExecuteInfo(jobSchedulePlan)
	scheduler.jobExcutingTable[jobSchedulePlan.Job.Name] = jobExecuteInfo

	// exec
	fmt.Println("执行任务：", jobExecuteInfo.Job.Name, jobExecuteInfo.PlanTime, jobExecuteInfo.RealTime)
	G_executor.ExecuteJob(jobExecuteInfo)
}

// 重新计算任务调度状态
func (scheduler *Scheduler) TrySchedule() (scheduleAfter time.Duration) {
	var (
		jobSchedulePlan *common.JobSchedulePlan
		now             time.Time
		latestTime      *time.Time
	)

	// 如果没有任务, 随便睡眠， 但也不要太长
	if len(scheduler.jobPlanTable) == 0 {
		scheduleAfter = 1 * time.Second
		return
	}

	now = time.Now()
	// 1. 便利所有任务
	// 挑选最短执行周期精确睡眠
	for _, jobSchedulePlan = range scheduler.jobPlanTable {
		if !jobSchedulePlan.NextTime.After(now) {
			// 2. 过期的任务立即执行
			scheduler.TryStartJob(jobSchedulePlan)
			fmt.Println("执行任务：", jobSchedulePlan.Job.Name, jobSchedulePlan.Job.CronExpr)
			jobSchedulePlan.NextTime = jobSchedulePlan.Expr.Next(now) // update next exec time
		}

		if latestTime == nil || jobSchedulePlan.NextTime.Before(*latestTime) {
			latestTime = &jobSchedulePlan.NextTime
		}
	}

	// 3。 统计最近的过期任务时间
	scheduleAfter = (*latestTime).Sub(now)
	return
}

// 调度协程
func (scheduler *Scheduler) schedulerLoop() {
	// 定时任务

	var (
		jobEvent      *common.JobEvent
		scheduleAfter time.Duration
		scheduleTimer *time.Timer
		jobResult     *common.JobExecuteResult
	)

	// 初始化一次内存中的job
	scheduleAfter = scheduler.TrySchedule()

	// 调度延时定时器
	scheduleTimer = time.NewTimer(scheduleAfter)

	for {
		select {
		case jobEvent = <-scheduler.jobEventChan: // 监听任务变化
			// 对内存中的jobPlan增删改查
			scheduler.handleJobEvent(jobEvent)
		case <-scheduleTimer.C: // 时间间隔到期
		case jobResult = <-scheduler.jobResultChan: // 监听任务执行结果
			scheduler.handleJobResult(jobResult)
		}
		scheduleAfter = scheduler.TrySchedule()
		scheduleTimer.Reset(scheduleAfter)
	}
}

// 推送任务事件
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	scheduler.jobEventChan <- jobEvent
}

// 推送任务执行结果
func (scheduler *Scheduler) PushJobResult(jobExecuteResult *common.JobExecuteResult) {
	scheduler.jobResultChan <- jobExecuteResult
}

func InitScheduler() (err error) {
	G_Scheduler = &Scheduler{
		jobEventChan:     make(chan *common.JobEvent, 1000),
		jobPlanTable:     make(map[string]*common.JobSchedulePlan),
		jobExcutingTable: make(map[string]*common.JobExecuteInfo),
		jobResultChan:    make(chan *common.JobExecuteResult, 1000),
	}

	go G_Scheduler.schedulerLoop()
	return
}
