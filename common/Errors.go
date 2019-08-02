package common

import "errors"

var (
	ERR_LOCK_ALRERADY_REQUIRED = errors.New("锁已被占用")

	ERR_NO_LOCAL_IP = errors.New("本机没有物理网卡")
)
