package coster

import (
	"time"

	"github.com/astaxie/beego/logs"
)

const configPath = "./config/coster.json"

var log *logs.BeeLogger

// 用于保存状态的配置结构
type costLogConfig struct {
	LastTime  time.Time // 上次使用的时间
	MonthCost float64   // 月度消费
	TotalCost float64   // 总消费
	LogsPath  string    // 日志保持位置
	WritePath string    // 消费历史保存位置
}

// 初始化的配置
var dfData = costLogConfig{
	LastTime:  time.Now(),
	MonthCost: 0.0,
	TotalCost: 0.0,
	WritePath: "./data/coster/coster.txt",
}
