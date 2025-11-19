package requests

import (
	"sync/atomic"
)

// 使用原子操作保证并发安全
var debugMode int32 = 0

// SetDebug 开启或关闭 Debug 模式
// enable: true 开启, false 关闭
func SetDebug(enable bool) {
	if enable {
		atomic.StoreInt32(&debugMode, 1)
	} else {
		atomic.StoreInt32(&debugMode, 0)
	}
}

// IsDebug 判断当前是否处于 Debug 模式
func IsDebug() bool {
	return atomic.LoadInt32(&debugMode) == 1
}
