package worker

import (
	"context"
	"github.com/Mopip77/crontab/common"
	"go.etcd.io/etcd/clientv3"
)

type JobLock struct {
	kv         clientv3.KV
	lease      clientv3.Lease
	jobName    string
	cancelFunc context.CancelFunc
	leaseId    clientv3.LeaseID
	isLocked   bool
}

func InitJobLock(jobName string, kv clientv3.KV, lease clientv3.Lease) (jobLock *JobLock) {
	jobLock = &JobLock{
		kv:      kv,
		lease:   lease,
		jobName: jobName,
	}
	return
}

func (jobLock *JobLock) TryLock() (err error) {
	var (
		leaseGrantResp *clientv3.LeaseGrantResponse
		cancelCtx      context.Context
		cancelFunc     context.CancelFunc
		leaseId        clientv3.LeaseID
		leaseKeepChan  <-chan *clientv3.LeaseKeepAliveResponse
		txn            clientv3.Txn
		lockKey        string
		txnResp        *clientv3.TxnResponse
	)
	// 1.创建租约
	if leaseGrantResp, err = jobLock.lease.Grant(context.TODO(), 5); err != nil {
		return
	}

	cancelCtx, cancelFunc = context.WithCancel(context.TODO())

	// 租约id
	leaseId = leaseGrantResp.ID

	// 2.自动续租
	if leaseKeepChan, err = jobLock.lease.KeepAlive(cancelCtx, leaseId); err != nil {
		goto FAIL
	}

	// 3. 处理续租应答的协程
	go func() {
		var (
			keepResp *clientv3.LeaseKeepAliveResponse
		)

		for {
			select {
			case keepResp = <-leaseKeepChan:
				if keepResp == nil {
					goto END
				}
			}
		}
	END:
	}()

	// 4. 创建事物txn
	txn = jobLock.kv.Txn(context.TODO())

	// 锁路径
	lockKey = common.JOB_LOCK_DIR + jobLock.jobName

	// 5. 抢锁
	txn.If(clientv3.Compare(clientv3.CreateRevision(lockKey), "=", 0)).
		Then(clientv3.OpPut(lockKey, "", clientv3.WithLease(leaseId))).
		Else(clientv3.OpGet(lockKey))

	// 提交事物
	if txnResp, err = txn.Commit(); err != nil {
		goto FAIL
	}

	// 6. 成功返回， 失败释放租约
	if !txnResp.Succeeded {
		err = common.ERR_LOCK_ALRERADY_REQUIRED
		goto FAIL
	}

	// 记录租约id, cancelFunc
	jobLock.leaseId = leaseId
	jobLock.cancelFunc = cancelFunc
	jobLock.isLocked = true

	return

FAIL:
	cancelFunc()
	jobLock.lease.Revoke(context.TODO(), leaseId)
	return
}

func (jobLock *JobLock) UnLock() (err error) {
	if jobLock.isLocked {
		jobLock.cancelFunc()
		_, err = jobLock.lease.Revoke(context.TODO(), jobLock.leaseId)
	}
	return
}
