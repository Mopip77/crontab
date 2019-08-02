package worker

import (
	"github.com/Mopip77/crontab/common"
	"os/exec"
	"time"
)

type Executor struct {
}

var (
	G_executor *Executor
)

// 执行一个job
func (executor *Executor) ExecuteJob(info *common.JobExecuteInfo) {
	go func() {
		var (
			cmd     *exec.Cmd
			err     error
			output  []byte
			result  *common.JobExecuteResult
			jobLock *JobLock
		)

		result = &common.JobExecuteResult{
			ExecuteInfo: info,
			StartTime:   time.Now(),
		}

		// 初始化分布式锁
		jobLock = G_jobMgr.CreateJobLock(info.Job.Name)
		// 抢锁
		err = jobLock.TryLock()
		defer jobLock.UnLock()
		if err != nil { // 上锁失败
			result.Err = err
			result.EndTime = time.Now()
		} else {
			// TryLock 是网络操作， 所以startTime需要被重置
			result.StartTime = time.Now()
			// 执行shell命令
			cmd = exec.CommandContext(info.CancelCtx, "/bin/bash", "-c", info.Job.Command)
			output, err = cmd.CombinedOutput()

			result.Output = output
			result.Err = err
			result.EndTime = time.Now()

		}
		// 推送执行结果
		G_Scheduler.PushJobResult(result)
	}()
}

func InitExecutor() (err error) {
	G_executor = &Executor{}
	return
}
