package master

import (
	"context"
	"github.com/Mopip77/crontab/common"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type LogMgr struct {
	client        *mongo.Client
	logCollection *mongo.Collection
}

var (
	G_LogMgr *LogMgr
)

func InitLogMgr() (err error) {
	var (
		client *mongo.Client
	)

	if client, err = mongo.Connect(context.TODO(),
		options.Client().ApplyURI(G_config.MongoUri),
		options.Client().SetConnectTimeout(time.Duration(G_config.MongoTimeout)*time.Millisecond)); err != nil {
		return
	}

	G_LogMgr = &LogMgr{
		client:        client,
		logCollection: client.Database("cron").Collection("log"),
	}

	return
}

func (logMgr *LogMgr) ListLog(name string, skip int, limit int) (logArr []*common.JobLog, err error) {
	var (
		filter  *common.JobLogFilter
		logSort *common.SortLogByStartTime
		cursor  *mongo.Cursor
		jobLog  *common.JobLog
	)

	// 避免没有log， 返回空指针
	logArr = make([]*common.JobLog, 0)

	filter = &common.JobLogFilter{
		JobName: name,
	}

	// 倒序
	logSort = &common.SortLogByStartTime{
		SortOrder: -1,
	}

	// 查询
	if cursor, err = logMgr.logCollection.Find(context.TODO(), filter, options.Find().SetSort(logSort), options.Find().SetSkip(int64(skip)), options.Find().SetLimit(int64(limit))); err != nil {
		return
	}

	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		jobLog = &common.JobLog{}

		if err = cursor.Decode(jobLog); err != nil {
			continue
		}

		logArr = append(logArr, jobLog)
	}
	return
}
