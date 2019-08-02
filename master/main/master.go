package main

import (
	"flag"
	"fmt"
	"github.com/Mopip77/crontab/master"
	"runtime"
	"time"
)

var (
	configFile string // config file path
)

// 解析命令行参数
func initArgs() {
	flag.StringVar(&configFile, "config", "./master.json", "指定master.json")
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
	if err = master.InitConfig(configFile); err != nil {
		goto ERR
	}

	if err = master.InitWorkerMgr(); err != nil {
		goto ERR
	}

	if err = master.InitLogMgr(); err != nil {
		goto ERR
	}

	if err = master.InitJobMgr(); err != nil {
		goto ERR
	}

	// 启动HTTP Api服务
	if err = master.InitApiServer(); err != nil {
		goto ERR
	}

	for {
		time.Sleep(1 * time.Second)
	}

	return

ERR:
	fmt.Println(err)
}
