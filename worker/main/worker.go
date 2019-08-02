package main

import (
	"flag"
	"fmt"
	"github.com/Mopip77/crontab/worker"
	"runtime"
	"time"
)

var (
	configFile string // config file path
)

// 解析命令行参数
func initArgs() {
	flag.StringVar(&configFile, "config", "./worker.json", "指定worker.json")
	flag.Parse()
}

func initEnv() {
	// 设置go线程数量为cpu核心数量
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	var (
		err error
	)

	// 初始化命令行参数
	initArgs()

	// 初始化线程
	initEnv()

	// 加载配置
	if err = worker.InitConfig(configFile); err != nil {
		goto ERR
	}

	if err = worker.InitRegister(); err != nil {
		goto ERR
	}

	if err = worker.InitLogSink(); err != nil {
		goto ERR
	}

	if err = worker.InitExecutor(); err != nil {
		goto ERR
	}

	if err = worker.InitScheduler(); err != nil {
		goto ERR
	}

	if err = worker.InitJobMgr(); err != nil {
		goto ERR
	}

	for {
		time.Sleep(1 * time.Second)
	}

	return

ERR:
	fmt.Println(err)
}
