package eventer

// 事件记录器

import (
	"log"
	"time"
)

const configPath = "./config/eventer.json" // 配置文件路径

var (
	l *log.Logger
)

// 配置数据结构
type eventLogConfig struct {
	LastTime   time.Time // 上次写入的时间
	TodayTimes int       // 今日记录日志数量
	WritePath  string    // 日志文件存放路径
}

// 默认配置数据
var defaultConfig = eventLogConfig{
	LastTime:   time.Now(),
	TodayTimes: 0,
	WritePath:  "./data/eventer/itmonth.txt",
}
